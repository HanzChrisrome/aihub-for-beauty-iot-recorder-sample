package wsclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/recorder"
)

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {

	case MSG_START_RECORDING:
		c.handleStartRecording()

	case MSG_STOP_RECORDING:
		c.handleStopRecording()

	case MSG_STATUS:
		log.Println("?? Status from server:", string(msg.Data))

	case MSG_ERROR:
		log.Println("? Server error:", msg.Message)

	case MSG_SUCCESS:
		return

	default:
		return
	}
}

func (c *Client) handleStartRecording() {
	log.Println("?? START RECORDING tangina mo aeron received")

	if err := recorder.Start(); err != nil {
		c.sendError("Failed to start recording: " + err.Error())
		return
	}

	c.sendSuccess("recording started")
}

func (c *Client) handleStopRecording() {
	log.Println("?? STOP RECORDING received")

	if err := recorder.Stop(); err != nil {
		c.sendError("Failed to stop recording: " + err.Error())
		return
	}

	filePath := audio.GetLastFilePath()
	if err := c.sendFile(filePath); err != nil {
		c.sendError("Failed to send file: " + err.Error())
		return
	}

	c.sendSuccess("recording stopped and file sent")
}

func (c *Client) sendFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	log.Println(">> Sent audio file to backend:", filePath)
	return nil
}
