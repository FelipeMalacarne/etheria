package account

import "time"

// User represents a persisted account entity.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// PublicUser is the safe subset returned to the client.
type PublicUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
	}
}
