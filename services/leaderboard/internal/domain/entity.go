package domain

import (
	"errors"
)

var (
	ErrNotFound = errors.New("leaderboard: resource not found")
	ErrInvalid  = errors.New("leaderboard: invalid input")
)

type RankEntry struct {
	UserID       string    `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	Rank         int       `json:"rank"`
	RewardCount  int       `json:"reward_count"`
	FocusMinutes int       `json:"focus_minutes"`
	Period       string    `json:"period"`
}
