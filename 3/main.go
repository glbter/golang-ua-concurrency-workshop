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
	Expiry int
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
	}

	go m.cleanJob()
	return m
}

func (m *SessionManager) cleanJob() {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			m.cleanUp()
		}
	}
}

func (m *SessionManager) cleanUp() {
	m.mux.Lock()
	for k := range m.sessions {
		v, ok := m.sessions[k]
		if !ok {
			continue
		}

		v.Expiry++
		if v.Expiry >= 5 {
			delete(m.sessions, k)
			continue
		}

		m.sessions[k] = v
	}
	m.mux.Unlock()
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
	_, ok := m.sessions[sessionID]
	m.mux.RUnlock()
	if !ok {
		return ErrSessionNotFound
	}

	m.mux.Lock()
	m.sessions[sessionID] = Session{
		Data: data,
		Expiry: 0,
	}
	m.mux.Unlock()

	return nil
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