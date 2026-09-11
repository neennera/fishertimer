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
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Species   string    `json:"species"`
	Rarity    string    `json:"rarity"`
	AwardedAt time.Time `json:"awarded_at"`
}
