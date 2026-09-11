package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("account: resource not found")
	ErrInvalid  = errors.New("account: invalid input")
)

type Profile struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	IsBanned    bool      `json:"is_banned"`
	TotalFocus  int       `json:"total_focus_minutes"`
	UpdatedAt   time.Time `json:"updated_at"`
}
