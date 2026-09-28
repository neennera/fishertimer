package domain

import (
	"context"
	"time"
)

type Repository interface {
	Award(ctx context.Context, r *FishReward) error
	ListByUser(ctx context.Context, userID string) ([]FishReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
}

