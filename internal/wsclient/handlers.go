package wsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
	"github.com/otis-co-ltd/aihub-recorder/internal/recorder"
)

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {

	case MSG_START_RECORDING:
		// Try to parse as multi-device message first
		var multiMsg StartRecordingMessage
		if err := json.Unmarshal(msg.Data, &multiMsg); err == nil && multiMsg.SessionID != "" {
			c.handleStartRecordingMulti(multiMsg)
		}

	case MSG_STOP_RECORDING:
		// Try to parse as session-based message first
		var stopMsg StopRecordingMessage
		if err := json.Unmarshal(msg.Data, &stopMsg); err == nil && stopMsg.SessionID != "" {
			c.handleStopRecordingSession(stopMsg)
		}

	case MSG_LIST_DEVICES:
		c.handleListDevices()

	case MSG_STOP_ALL:
		c.handleStopAll()

	case MSG_STATUS:
		log.Println("📊 Status from server:", string(msg.Data))

	case MSG_ERROR:
		log.Println("❌ Server error:", msg.Message)

	case MSG_SUCCESS:
		return

	default:
		return
	}
}

// handleStartRecordingMulti handles multi-device recording
func (c *Client) handleStartRecordingMulti(msg StartRecordingMessage) {
	// If a device name is provided, resolve it to an index
	if msg.DeviceName != "" {
		idx, err := audio.GetDeviceIndexByName(msg.DeviceName)
		if err != nil {
			c.sendErrorMessage("start_recording", fmt.Sprintf("Failed to find device by name: %v", err))
			return
		}
		log.Printf("Resolved device name %q -> index %d", msg.DeviceName, idx)
		msg.DeviceIndex = idx
	}

	log.Printf("🎙️ Starting recording for session: %s, device: %d", msg.SessionID, msg.DeviceIndex)

	err := recorder.StartSession(msg.SessionID, msg.DeviceIndex)
	if err != nil {
		c.sendErrorMessage("start_recording", fmt.Sprintf("Failed to start recording: %v", err))
		return
	}

	c.sendSuccessMessage("start_recording", fmt.Sprintf("Recording started for session %s on device %d", msg.SessionID, msg.DeviceName))
}

// handleStopRecordingSession handles session-based stop
func (c *Client) handleStopRecordingSession(msg StopRecordingMessage) {
	log.Printf("?? Stopping recording for session: %s", msg.SessionID)

	filePath, err := recorder.StopSession(msg.SessionID)
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "not found") || strings.Contains(lower, "no such") || strings.Contains(lower, "no session") || strings.Contains(lower, "does not exist") {
			c.sendErrorMessage("stop_recording", fmt.Sprintf("no active session found with id %s", msg.SessionID))
		} else {
			c.sendErrorMessage("stop_recording", fmt.Sprintf("Failed to stop recording: %v", err))
		}
		return
	}

	go func() {
		finalPath := filePath
		if config.Load().SYS_ENABLE_DENOISING {
			log.Printf("?? Applying RNNoise denoising to: %s", filePath)
			denoisedPath, err := audio.DenoiseAudioFile(filePath)
			if err != nil {
				log.Printf("?? Denoising failed: %v, uploading original file", err)
				c.sendErrorMessage("denoise", fmt.Sprintf("Denoising failed: %v", err))
			} else {
				log.Printf("? Denoised audio saved to: %s", denoisedPath)
				finalPath = denoisedPath
			}
		}

		// Upload the final file (denoised or original)
		err := c.sendFile(finalPath, msg.SessionID)
		if err != nil {
			c.sendErrorMessage("upload_file", fmt.Sprintf("Failed to upload: %v", err))
		} else {
			c.sendSuccessMessage("upload_file", fmt.Sprintf("File uploaded for session %s", msg.SessionID))
		}
	}()

	c.sendSuccessMessage("stop_recording", fmt.Sprintf("Recording stopped for session %s", msg.SessionID))
}

// handleListDevices lists all available audio devices
func (c *Client) handleListDevices() {
	devices := audio.ListAudioDevices()

	response := ResponseMessage{
		Command: "list_devices_response",
		Status:  "success",
		Message: "Available audio devices",
		Data:    devices,
	}

	c.sendResponse(response)
}

// handleStopAll stops all active recording sessions
func (c *Client) handleStopAll() {
	fileMap, err := recorder.StopAllSessions()
	if err != nil {
		if len(fileMap) == 0 {
			c.sendErrorMessage("stop_all", fmt.Sprintf("Failed to stop sessions: %v", err))
			return
		}
		log.Printf("stop_all: some sessions failed to stop: %v", err)
		c.sendErrorMessage("stop_all", fmt.Sprintf("Some sessions failed to stop: %v", err))
	}

	for sessionID, filePath := range fileMap {
		if filePath == "" {
			c.sendErrorMessage("upload_file", fmt.Sprintf("no file produced for session %s", sessionID))
			continue
		}
		go func(sid, fp string) {
			if err := c.sendFile(fp, sid); err != nil {
				c.sendErrorMessage("upload_file", fmt.Sprintf("Failed to upload for %s: %v", sid, err))
			} else {
				c.sendSuccessMessage("upload_file", fmt.Sprintf("File uploaded for session %s", sid))
			}
		}(sessionID, filePath)
	}

	c.sendSuccessMessage("stop_all", "All recording sessions stopped")
}

// sendFile uploads a file to the backend with session information
func (c *Client) sendFile(filePath string, sessionID string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	// Add session_id field
	if err := writer.WriteField("session_id", sessionID); err != nil {
		return err
	}

	writer.Close()

	url := "http://aeronsarondo.site/db/audio"
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	log.Println("📤 Sent audio file to backend:", filePath, "Session:", sessionID)
	return nil
}
