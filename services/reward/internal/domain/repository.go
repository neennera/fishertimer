package domain

import (
	"context"
)

type Repository interface {
	Award(ctx context.Context, r *UnlockedReward) error
	ListByUser(ctx context.Context, userID string) ([]UnlockedReward, error)
}
