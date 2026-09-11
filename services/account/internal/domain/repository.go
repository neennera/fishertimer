package domain

import (
	"context"
)

type Repository interface {
	GetProfile(ctx context.Context, userID string) (*Profile, error)
	UpdateBanStatus(ctx context.Context, userID string, banned bool) error
}
