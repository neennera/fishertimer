package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("reward: resource not found")
	ErrInvalid  = errors.New("reward: invalid input")
)

type FishReward struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Species     string    `json:"species"`
	Rarity      string    `json:"rarity"`
	AwardedAt   time.Time `json:"awarded_at"`
}
