package auth

import (
	"context"
	"strings"
	"time"

	"github.com/felipemalacarne/etheria/internal/domain/account"
)

// PasswordHasher abstracts hashing and verification for passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// SessionStore manages auth tokens for logged-in users.
type SessionStore interface {
	Create(userID string) (string, error)
	Resolve(token string) (string, bool)
	Revoke(token string)
}

// IDGenerator creates unique identifiers for new users.
type IDGenerator interface {
	New() string
}

// Service coordinates account registration/login workflows.
type Service struct {
	repo     account.Repository
	hasher   PasswordHasher
	sessions SessionStore
	ids      IDGenerator
	clock    func() time.Time
}

func NewService(repo account.Repository, hasher PasswordHasher, sessions SessionStore, ids IDGenerator) *Service {
	return &Service{
		repo:     repo,
		hasher:   hasher,
		sessions: sessions,
		ids:      ids,
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *Service) Register(ctx context.Context, email, username, password string) (account.PublicUser, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	username = strings.TrimSpace(username)
	if email == "" || username == "" || len(password) < 6 {
		return account.PublicUser{}, "", account.ErrInvalidInput
	}

	if existing, ok, err := s.repo.FindByEmail(ctx, email); err != nil {
		return account.PublicUser{}, "", err
	} else if ok && existing.ID != "" {
		return account.PublicUser{}, "", account.ErrEmailExists
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return account.PublicUser{}, "", err
	}

	user := account.User{
		ID:           s.ids.New(),
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    s.clock(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if err == account.ErrEmailExists {
			return account.PublicUser{}, "", err
		}
		return account.PublicUser{}, "", err
	}

	token, err := s.sessions.Create(user.ID)
	if err != nil {
		return account.PublicUser{}, "", err
	}

	return user.Public(), token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (account.PublicUser, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return account.PublicUser{}, "", account.ErrInvalidCredentials
	}

	user, ok, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return account.PublicUser{}, "", err
	}
	if !ok {
		return account.PublicUser{}, "", account.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return account.PublicUser{}, "", account.ErrInvalidCredentials
	}

	token, err := s.sessions.Create(user.ID)
	if err != nil {
		return account.PublicUser{}, "", err
	}

	return user.Public(), token, nil
}

func (s *Service) AuthenticateToken(ctx context.Context, token string) (account.User, bool, error) {
	if token == "" {
		return account.User{}, false, nil
	}

	userID, ok := s.sessions.Resolve(token)
	if !ok {
		return account.User{}, false, nil
	}

	user, found, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return account.User{}, false, err
	}
	if !found {
		return account.User{}, false, nil
	}

	return user, true, nil
}

func (s *Service) Revoke(token string) {
	if token == "" {
		return
	}
	s.sessions.Revoke(token)
}
