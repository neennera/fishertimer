package domain

import (
	"context"
)

type Repository interface {
	Award(ctx context.Context, r *FishReward) error
	ListByUser(ctx context.Context, userID string) ([]FishReward, error)
}
