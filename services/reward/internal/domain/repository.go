package domain

import (
	"context"
	"time"
)

type Repository interface {
	// ListItems returns the reward_items catalogue.
	ListItems(ctx context.Context) ([]RewardItem, error)
	// ListByCycle returns the rewards stored for a cycle_id.
	ListByCycle(ctx context.Context, cycleID string) ([]UnlockedReward, error)
	// AwardMany stores one user_rewards row per reward, filling in IDs and
	// awarded_at. rewards[i] is draw i of its cycle. Returns
	// ErrAlreadyAwarded if the cycle's rewards are already stored.
	AwardMany(ctx context.Context, rewards []UnlockedReward) error
	ListByUser(ctx context.Context, userID string) ([]UnlockedReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
}
