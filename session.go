package session

import (
	"fmt"
	"net/http"

	"github.com/gochef/chef/utils"
	"github.com/gochef/cookie"
)

type (
	// Store represents an interface to control session object
	Store interface {
		// Get returns an item saved in session
		Get(key string) (interface{}, bool)
		Set(key string, data interface{})
		Remove(key string)
		Clear()
		ID() string
	}

	// Provider represents a session provider interface
	Provider interface {
		Read(sid string, expires int64) Store
		Initialize(sid string, expires int64) Store
		Exists(sid string) bool
		Regenerate(oldsid string, newsid string) Store
		Destroy(sid string)
	}

	// Session represents a single session instance
	Session struct {
		id       string
		provider Provider
		config   *Config
		store    Store
	}

	// Config is the session instance configuration
	Config struct {
		Provider     string
		Key          string
		CookieLength int
		MaxAge       int64
	}
)

var (
	providers = map[string]Provider{
		"memory": MemoryProvider,
	}
)

// New returns a session instance with configured provider
func New(cfg *Config) *Session {
	provider, ok := providers[cfg.Provider]
	if !ok {
		errStr := "Session Provider %s is not registered"
		panic(fmt.Sprintf(errStr, provider))
	}

	return &Session{
		provider: provider,
		config:   cfg,
	}
}

// Start starts a session instance
func (s *Session) Start(w http.ResponseWriter, req *http.Request) {
	cookieValue := cookie.Get(s.config.Key, req)

	if cookieValue == "" { //Empty session cookie //Start new session
		s.id, _ = utils.RandomString(s.config.CookieLength)
		s.store = s.provider.Initialize(s.id, s.config.MaxAge)

		ck := cookie.AcquireCookie()
		ck.Name = s.config.Key
		ck.Value = s.id
		ck.HttpOnly = true
		ck.MaxAge = int(s.config.MaxAge)

		cookie.Add(ck, w)
		cookie.ReleaseCookie(ck)
	} else {
		s.store = s.provider.Read(cookieValue, s.config.MaxAge)
	}
}

// RegisterProvider adds a provider to usable list.
// panics if provider is already registered
func RegisterProvider(providerName string, provider Provider) {
	if _, ok := providers[providerName]; ok {
		errStr := fmt.Sprintf("session: Provider %s is already registered", providerName)
		panic(errStr)
	}

	providers[providerName] = provider
}

// Get fetches an item from session store by key,
// returns an empty interface and false if it doesnt exist
func (s *Session) Get(key string) (interface{}, bool) {
	return s.store.Get(key)
}

// GetString returns a string item from session store
func (s *Session) GetString(key string) (string, bool) {
	data, ok := s.Get(key)
	if !ok {
		return "", false
	}

	str, ok := data.(string)
	return str, ok
}

// GetInt returns an integer item from session store
func (s *Session) GetInt(key string) (int, bool) {
	data, ok := s.Get(key)
	if !ok {
		return 0, false
	}

	i, ok := data.(int)
	return i, ok
}

// Set adds an item to session store, identified by provided key
func (s *Session) Set(key string, data interface{}) {
	s.store.Set(key, data)
}

// Remove deletes an item from session store by provided key
func (s *Session) Remove(key string) {
	s.store.Remove(key)
}

// Pull gets an item from session store and deletes the item from session
func (s *Session) Pull(key string) (interface{}, bool) {
	data, ok := s.store.Get(key)
	s.Remove(key)

	return data, ok
}

// PullString gets a string item from session store and deletes the item from session
func (s *Session) PullString(key string) (string, bool) {
	data, ok := s.GetString(key)
	s.Remove(key)

	return data, ok
}

// PullInt gets an integer item from session store and deletes the item from session
func (s *Session) PullInt(key string) (int, bool) {
	data, ok := s.GetInt(key)
	s.Remove(key)

	return data, ok
}

// Clear empties the session store
func (s *Session) Clear() {
	s.store.Clear()
}

// ID returns the session id
func (s *Session) ID() string {
	return s.id
}
