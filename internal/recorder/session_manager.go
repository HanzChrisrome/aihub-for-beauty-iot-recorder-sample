package recorder

import (
	"fmt"
	"sync"
)

type SessionManager struct {
	sessions map[string]*RecordingSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*RecordingSession),
	}
}

func (sm *SessionManager) CreateSession(sessionID string, deviceIndex int) (*RecordingSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session already exists
	if _, exists := sm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session with ID %s already exists", sessionID)
	}

	// Create new session
	session := NewRecordingSession(sessionID, deviceIndex)
	sm.sessions[sessionID] = session
	return session, nil
}

func (sm *SessionManager) GetSession(sessionID string) (*RecordingSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", sessionID)
	}
	return session, nil
}

func (sm *SessionManager) RemoveSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(sm.sessions, sessionID)
	return nil
}

func (sm *SessionManager) GetAllSessions() []*RecordingSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*RecordingSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SessionManager) GetActiveCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
