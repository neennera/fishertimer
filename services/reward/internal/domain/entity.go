package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("reward: resource not found")
	ErrInvalid  = errors.New("reward: invalid input")
)

// Categories, matching reward_items.category's Mongo schema validator.
const (
	CategoryFish       = "FISH"
	CategoryDecoration = "DECORATION"
	CategoryRod        = "ROD"
)

// Rarity tiers, matching reward_items.rarity's Mongo schema validator.
const (
	RarityCommon    = "COMMON"
	RarityUncommon  = "UNCOMMON"
	RarityRare      = "RARE"
	RarityEpic      = "EPIC"
	RarityLegendary = "LEGENDARY"
)

// UnlockedReward is one catch/drop: a user_rewards row joined with its
// reward_items catalog details
type UnlockedReward struct {
	UserRewardID string    `json:"user_reward_id"`
	ItemID       string    `json:"item_id"`
	UserID       string    `json:"user_id"`
	CycleID      string    `json:"cycle_id"`
	ItemName     string    `json:"item_name"`
	Category     string    `json:"category"`
	Rarity       string    `json:"rarity"`
	BaseWeight   float64   `json:"base_weight"`
	ScoreValue   int       `json:"score_value"`
	AssetURL     string    `json:"asset_url"`
	AwardedAt    time.Time `json:"awarded_at"`
}
