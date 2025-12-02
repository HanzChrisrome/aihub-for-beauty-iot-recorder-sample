package recorder

import (
	"sync"
	"time"

	"github.com/otis-co-ltd/aihub-recorder/internal/audio"
)

type RecordingSession struct {
	SessionID string
	DeviceIndex int 
	Recorder audio.IAudioFormat
	Control *audio.RecondControlSignal
	StartTime time.Time
	FilePath string
	mu sync.Mutex
	isRecording bool
}

func NewRecordingSession(sessionID string, deviceIndex int) *RecordingSession {
	return &RecordingSession{
		SessionID: sessionID,
		DeviceIndex: deviceIndex,
		Control: audio.NewRecControlSig(),
		StartTime: time.Now(),
		isRecording: false,
	}
}

func (s *RecordingSession) IsRecording() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRecording
}

func (s *RecordingSession) SetRecording(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRecording = state
}

func (s *RecordingSession) SetFilePath(path string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.FilePath = path
}

func (s *RecordingSession) GetFilePath() string {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.FilePath
}