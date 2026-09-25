package domain

import (
	"context"
)

// Repository is the driven port for account_db persistence. It must never leak
// SQL types: adapters in internal/adapter/repository implement it against
// PostgreSQL or an in-memory map.
type Repository interface {
	CreateUser(ctx context.Context, u *UserAccount) error
	GetUserByEmail(ctx context.Context, email string) (*UserAccount, error)
	GetUserByID(ctx context.Context, userID string) (*UserAccount, error)
}
