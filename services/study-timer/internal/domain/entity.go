package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-timer: resource not found")
	ErrInvalid  = errors.New("study-timer: invalid input")
)

type TimerPhase string

const (
	PhaseWork TimerPhase = "WORK"
	PhaseRest TimerPhase = "REST"
)

type TimerState struct {
	SessionID    string     `json:"session_id"`
	UserID       string     `json:"user_id"`
	Status       string     `json:"status"`
	Phase        TimerPhase `json:"phase"`
	WorkMinutes  int        `json:"work_minutes"`
	RestMinutes  int        `json:"rest_minutes"`
	CurrentCycle int        `json:"current_cycle"`
	LastUpdated  time.Time  `json:"last_updated"`
}

type TimerHistory struct {
	UserID            string       `json:"user_id"`
	SessionsJoined    int          `json:"sessions_joined"`
	CyclesCompleted   int          `json:"cycles_completed"`
	TotalFocusMinutes int          `json:"total_focus_minutes"`
	LastActive        time.Time    `json:"last_active"`
	DailyFocusMinutes []DailyFocus `json:"daily_focus_minutes"`
}


type DailyFocus struct {
	Date         string `json:"date"` 
	FocusMinutes int    `json:"focus_minutes"`
}
