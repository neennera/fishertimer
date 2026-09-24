package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("account: resource not found")
	ErrInvalid  = errors.New("account: invalid input")
)

type UserAccount struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Profile struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserStatistics struct {
	UserID        string `json:"user_id"`
	TotalSessions int    `json:"total_sessions"`
	TotalFocusMin int    `json:"total_focus_minutes"`
	RewardsEarned int    `json:"rewards_earned"`
}
