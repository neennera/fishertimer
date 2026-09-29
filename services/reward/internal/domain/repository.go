package domain

import (
	"context"
	"time"
)

type Repository interface {
	Award(ctx context.Context, r *UnlockedReward) error
	ListByUser(ctx context.Context, userID string) ([]UnlockedReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
}
