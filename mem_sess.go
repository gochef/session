package session

import (
	"sync"
	"time"
)

type (
	// MemorySessionStore represents a session store
	MemorySessionStore struct {
		sid            string
		lastAccessedAt int64
		expresAt       int64
		values         map[string]interface{}
		sync.RWMutex
	}
)

// MemoryProvider is a variable holding the memory session provider
var MemoryProvider = &MemorySessionProvider{
	sessions: make(map[string]*MemorySessionStore),
}

// Get fetches an item from the session
// returns a boolean that indicates whether the item was found or not
func (s *MemorySessionStore) Get(key string) (interface{}, bool) {
	data, ok := s.values[key]
	return data, ok
}

// Set puts an item into the session
func (s *MemorySessionStore) Set(key string, data interface{}) {
	s.Lock()
	s.values[key] = data
	s.Unlock()
}

// Remove removes an item from the session
func (s *MemorySessionStore) Remove(key string) {
	s.RLock()
	delete(s.values, key)
	s.RUnlock()
}

// ID returns the session ID
func (s *MemorySessionStore) ID() string {
	return s.sid
}

// Clear empties the session
func (s *MemorySessionStore) Clear() {
	s.Lock()
	s.values = make(map[string]interface{})
	s.Unlock()
}

// MemorySessionProvider represents a MemorySession Provider instance
type MemorySessionProvider struct {
	maxAge   int64
	sessions map[string]*MemorySessionStore
	sync.RWMutex
}

// Read returns a MemorySessionStore
// If the Session store does not exist, a new one is created and returned
func (m *MemorySessionProvider) Read(sid string, maxAge int64) Store {
	m.RLock()

	if session, ok := m.sessions[sid]; ok {
		go m.Update(sid)
		m.RUnlock()
		return session
	}
	m.RUnlock()
	return m.Initialize(sid, maxAge)
}

// Initialize creates and returns a new MemorySessionStore
func (m *MemorySessionProvider) Initialize(sid string, maxAge int64) Store {
	m.Lock()

	m.maxAge = maxAge
	session := &MemorySessionStore{
		sid:            sid,
		lastAccessedAt: time.Now().Unix(),
		values:         make(map[string]interface{}),
	}

	m.sessions[sid] = session
	m.Unlock()
	return session
}

// Regenerate regenerates session
func (m *MemorySessionProvider) Regenerate(oldsid string, sid string) Store {
	if session, ok := m.sessions[oldsid]; ok {
		go m.Update(oldsid)
		session.sid = sid
		m.sessions[sid] = session
		delete(m.sessions, oldsid)

		return session
	}

	return m.Initialize(sid, m.maxAge)
}

// Exists checks if a session with passed id exists
func (m *MemorySessionProvider) Exists(sid string) bool {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.sessions[sid]; ok {
		return true
	}

	return false
}

// Update updates a session
func (m *MemorySessionProvider) Update(sid string) {
	m.Lock()
	defer m.Unlock()

	if session, ok := m.sessions[sid]; ok {
		session.lastAccessedAt = time.Now().Unix()
	}
}

// Destroy flushes the session
func (m *MemorySessionProvider) Destroy(sid string) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.sessions[sid]; ok {
		delete(m.sessions, sid)
	}
}
