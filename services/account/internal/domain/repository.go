package domain

import (
	"context"
)

type Repository interface {
	CreateUser(ctx context.Context, u *UserAccount) error
	GetUserByEmail(ctx context.Context, email string) (*UserAccount, error)
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	UpdateProfile(ctx context.Context, p *Profile) error
}
