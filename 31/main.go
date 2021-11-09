package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions map[string]Session
	mux sync.RWMutex
}

// Session stores the session's data
type Session struct {
	Data map[string]interface{}
	Timer *time.Timer
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
	}

	return m
}


// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.mux.Lock()
	m.sessions[sessionID] = Session{
		Data: make(map[string]interface{}),
		Timer: time.AfterFunc(5*time.Second, m.deleteSession(sessionID)),
	}
	m.mux.Unlock()

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.mux.RLock()
	session, ok := m.sessions[sessionID]
	m.mux.RUnlock()
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.mux.RLock()
	oldSession, ok := m.sessions[sessionID]
	m.mux.RUnlock()
	if !ok {
		return ErrSessionNotFound
	}

	oldSession.Timer.Reset(5*time.Second)
	m.mux.Lock()
	m.sessions[sessionID] = Session{
		Data: data,
		Timer: oldSession.Timer,
	}
	m.mux.Unlock()

	return nil
}

func (m *SessionManager) deleteSession(sessionId string) func() {
	return func() {
		m.mux.Lock()
		delete(m.sessions, sessionId)
		m.mux.Unlock()
	}
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	if err = m.UpdateSessionData(sID, data); err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}

// MakeSessionID is used to generate a random dummy sessionID
func MakeSessionID() (string, error) {
	buf := make([]byte, 26)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}