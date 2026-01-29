package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/felipemalacarne/etheria/internal/domain/account"
)

// UserRepository persists users in a JSON file (dev convenience).
type UserRepository struct {
	mu         sync.RWMutex
	path       string
	users      map[string]account.User
	emailIndex map[string]string
}

type filePayload struct {
	Users []account.User `json:"users"`
}

func NewUserRepository(path string) (*UserRepository, error) {
	repo := &UserRepository{
		path:       path,
		users:      make(map[string]account.User),
		emailIndex: make(map[string]string),
	}

	if err := repo.load(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *UserRepository) Create(_ context.Context, user account.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailKey := normalizeEmail(user.Email)
	if _, exists := r.emailIndex[emailKey]; exists {
		return account.ErrEmailExists
	}

	r.users[user.ID] = user
	r.emailIndex[emailKey] = user.ID

	if err := r.saveLocked(); err != nil {
		delete(r.users, user.ID)
		delete(r.emailIndex, emailKey)
		return err
	}

	return nil
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (account.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.emailIndex[normalizeEmail(email)]
	if !ok {
		return account.User{}, false, nil
	}

	user, exists := r.users[id]
	return user, exists, nil
}

func (r *UserRepository) GetByID(_ context.Context, id string) (account.User, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	return user, ok, nil
}

func (r *UserRepository) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var payload filePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	for _, user := range payload.Users {
		r.users[user.ID] = user
		r.emailIndex[normalizeEmail(user.Email)] = user.ID
	}

	return nil
}

func (r *UserRepository) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}

	payload := filePayload{Users: make([]account.User, 0, len(r.users))}
	ids := make([]string, 0, len(r.users))
	for id := range r.users {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		payload.Users = append(payload.Users, r.users[id])
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpPath, r.path)
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
