package domain

import "context"

type FishReward struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Species string `json:"species"`
	Rarity  string `json:"rarity"`
}

type RewardClient interface {
	ViewRewards(ctx context.Context, userID string) ([]FishReward, error)
}
