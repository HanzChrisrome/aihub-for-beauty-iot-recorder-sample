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

	default:
		return "aiff"
	}
}

func StartSession(sessionID string, deviceIndex int) error {
	// [STEP 1] Check if session already exists
	if _, err := sessionManager.GetSession(sessionID); err == nil {
		return fmt.Errorf("session %s already recording", sessionID)
	}

	// [STEP 2] Create new session
	session, err := sessionManager.CreateSession(sessionID, deviceIndex)
	if err != nil {
		return err
	}

	// [STEP 3] Create audio recorder instance
	audioTypeStr := getAudioTypeString(cfg.SYS_AUDIO_TYPE)
	session.Recorder = audio.NewAudioInstance(audioTypeStr)

	// [STEP 4] Set the microphone index BEFORE initializing
	session.Recorder.SetDeviceIndex(deviceIndex)

	// [STEP 5] Create session-specific directory
	sessionDir := filepath.Join(cfg.SYS_RECORD_PATH, sessionID)

	// [STEP 6] Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("device_%d_%s", deviceIndex, timestamp)

	expectedFilePath := filepath.Join(sessionDir, fmt.Sprintf("%s.%s", filename, audioTypeStr))
	session.SetFilePath(expectedFilePath)

	// [STEP 7] Initialize recorder with device index
	session.Recorder.Init(
		session.Control,
		sessionDir,
		filename,
		int16(cfg.SYS_AUDIO_CHANNEL),
		float64(cfg.SYS_AUDIO_SAMPLE_RATE),
		int(cfg.SYS_AUDIO_INPUT_BUFFER_SIZE),
	)

	// [STEP 8] Start recording in a separate goroutine
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

func StopAllSessions() (map[string]string, error) {
	sm := GetSessionManager()

	sm.mu.RLock()
	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	sm.mu.RUnlock()

	results := make(map[string]string)
	var lastErr error

	for _, id := range ids {
		fp, err := StopSession(id)
		if err != nil {
			lastErr = fmt.Errorf("stop %s: %w", id, err)
			continue
		}
		results[id] = fp
	}

	return results, lastErr
}

func GetActiveSessionCount() int {
	return sessionManager.GetActiveCount()
}

// GetSessionInfo returns information about a specific session
func GetSessionInfo(sessionID string) (*RecordingSession, error) {
	return sessionManager.GetSession(sessionID)
}
