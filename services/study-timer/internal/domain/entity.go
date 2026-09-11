package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-timer: resource not found")
	ErrInvalid  = errors.New("study-timer: invalid input")
)

type TimerState struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	WorkMinutes  int       `json:"work_minutes"`
	RestMinutes  int       `json:"rest_minutes"`
	CurrentCycle int       `json:"current_cycle"`
	LastUpdated  time.Time `json:"last_updated"`
}
