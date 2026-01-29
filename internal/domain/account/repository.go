package account

import "context"

// Repository defines persistence operations for accounts.
type Repository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, bool, error)
	GetByID(ctx context.Context, id string) (User, bool, error)
}
