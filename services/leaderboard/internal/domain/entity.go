package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("leaderboard: resource not found")
	ErrInvalid  = errors.New("leaderboard: invalid input")
)

// RankEntry is a single row in the leaderboard for a given period.
type RankEntry struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Rank        int    `json:"rank"`
	RewardCount int    `json:"reward_count"`
	Period      string `json:"period"`
}

// CachedRanking wraps a ranked list with its cache metadata.
type CachedRanking struct {
	Rankings         []RankEntry `json:"rankings"`
	Period           string      `json:"period"`
	LeaderboardFetch time.Time   `json:"leaderboard_last_fetch"`
	Cached           bool        `json:"cached"`
}
