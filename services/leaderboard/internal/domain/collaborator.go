package domain

import (
	"context"
	"time"
)

// FishReward mirrors the reward service's FishReward entity.
// Leaderboard only needs these fields to compute rankings.
type FishReward struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Species     string    `json:"species"`
	Rarity      string    `json:"rarity"`
	AwardedAt   time.Time `json:"awarded_at"`
}

// RewardClient is the outbound port to the Reward microservice.
type RewardClient interface {
	// ViewRewards fetches all rewards for a single user.
	ViewRewards(ctx context.Context, userID string) ([]FishReward, error)
	// ViewAllRewards fetches all rewards across all users for ranking computation.
	ViewAllRewards(ctx context.Context) ([]FishReward, error)
	// GetLastUpdate returns the timestamp of the most recent reward mutation.
	// Used by the S-1 cache-validation subflow.
	GetLastUpdate(ctx context.Context) (time.Time, error)
}
