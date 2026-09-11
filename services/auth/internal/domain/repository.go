package domain

import (
	"context"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*UserAccount, error)
	Create(ctx context.Context, user *UserAccount) error
}
