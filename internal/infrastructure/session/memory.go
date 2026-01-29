package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// MemoryStore is an in-memory implementation of auth.SessionStore.
type MemoryStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tokens: make(map[string]string),
	}
}

func (s *MemoryStore) Create(userID string) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.tokens[token] = userID
	s.mu.Unlock()

	return token, nil
}

func (s *MemoryStore) Resolve(token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, ok := s.tokens[token]
	return userID, ok
}

func (s *MemoryStore) Revoke(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
