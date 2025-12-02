package recorder

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
)

var sessionManager *SessionManager
var cfg *config.Config

func init() {
	cfg = config.Load()
	sessionManager = NewSessionManager()
}

func GetSessionManager() *SessionManager {
	return sessionManager
}

func getAudioTypeString(audioType uint8) string {
	switch audioType {
	case 0:
		return "aiff"
	case 1:
		return "wav"
	// Add other types as needed
	default:
		return "aiff" // Default to AIFF
	}
}

// StartSession starts recording for a specific session and device
func StartSession(sessionID string, deviceIndex int) error {
	// Check if session already exists
	if _, err := sessionManager.GetSession(sessionID); err == nil {
		return fmt.Errorf("session %s already recording", sessionID)
	}

	// Create new session
	session, err := sessionManager.CreateSession(sessionID, deviceIndex)
	if err != nil {
		return err
	}

	// Create audio recorder instance
	audioTypeStr := getAudioTypeString(cfg.SYS_AUDIO_TYPE)
	session.Recorder = audio.NewAudioInstance(audioTypeStr)

	// Set the device index BEFORE initializing
	session.Recorder.SetDeviceIndex(deviceIndex)

	// Create session-specific directory
	sessionDir := filepath.Join(cfg.SYS_RECORD_PATH, sessionID)

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("device_%d_%s", deviceIndex, timestamp)

	// Store the expected file path
	expectedFilePath := filepath.Join(sessionDir, fmt.Sprintf("%s.%v", filename, cfg.SYS_AUDIO_TYPE))
	session.SetFilePath(expectedFilePath)

	// Initialize recorder with device index
	session.Recorder.Init(
		session.Control,
		sessionDir,
		filename,
		int16(cfg.SYS_AUDIO_CHANNEL),
		float64(cfg.SYS_AUDIO_SAMPLE_RATE),
		int(cfg.SYS_AUDIO_INPUT_BUFFER_SIZE),
	)

	// Start recording in goroutine
	go session.Recorder.Record()
	session.SetRecording(true)

	return nil
}

// StopSession stops recording for a specific session
func StopSession(sessionID string) (string, error) {
	session, err := sessionManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	if !session.IsRecording() {
		return "", fmt.Errorf("session %s is not recording", sessionID)
	}

	// Send stop signal
	session.Control.Sig <- audio.AUDIO_CTL_STOP_REC

	// Wait for confirmation
	<-session.Control.Sig

	session.SetRecording(false)

	// Get file path before removing session
	filePath := session.FilePath

	// Remove session from manager
	sessionManager.RemoveSession(sessionID)

	return filePath, nil
}

func StopAllSessions() error {
	sessions := sessionManager.GetAllSessions()

	for _, session := range sessions {
		if session.IsRecording() {
			_, err := StopSession(session.SessionID)
			if err != nil {
				return fmt.Errorf("failed to stop session %s: %w", session.SessionID, err)
			}
		}
	}

	return nil
}

func GetActiveSessionCount() int {
	return sessionManager.GetActiveCount()
}

// GetSessionInfo returns information about a specific session
func GetSessionInfo(sessionID string) (*RecordingSession, error) {
	return sessionManager.GetSession(sessionID)
}
