package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-timer: resource not found")
	ErrInvalid  = errors.New("study-timer: invalid input")
	// ErrInvalidState is returned when an action is not allowed in the
	// timer's current status, e.g. pausing a timer that is not running.
	ErrInvalidState = errors.New("study-timer: action not allowed in current timer state")
)

type TimerPhase string

const (
	PhaseWork TimerPhase = "WORK"
	PhaseRest TimerPhase = "REST"
)

type TimerStatus string

const (
	StatusStopped TimerStatus = "STOPPED"
	StatusRunning TimerStatus = "RUNNING"
	StatusPaused  TimerStatus = "PAUSED"
)

const (
	DefaultWorkMinutes = 25
	DefaultRestMinutes = 5
)

// TimerState is one participant's timer in one room (UC-05).
//
// Progress is derived from server-recorded timestamps rather than from client
// ticks, so a disconnect or reload never alters the timer: time elapsed in the
// current phase is Elapsed plus, while running, the time since RunningSince.
type TimerState struct {
	SessionID    string
	UserID       string
	Status       TimerStatus
	Phase        TimerPhase
	WorkMinutes  int
	RestMinutes  int
	CurrentCycle int
	// RunningSince is when the timer last started or resumed. Only meaningful
	// while Status is RUNNING.
	RunningSince time.Time
	// Elapsed is the time already spent in the current phase before
	// RunningSince, i.e. the total of every earlier running stretch.
	Elapsed     time.Duration
	LastUpdated time.Time
}

// NewTimer returns a stopped timer with the default work and rest lengths.
func NewTimer(sessionID, userID string) *TimerState {
	return &TimerState{
		SessionID:   sessionID,
		UserID:      userID,
		Status:      StatusStopped,
		Phase:       PhaseWork,
		WorkMinutes: DefaultWorkMinutes,
		RestMinutes: DefaultRestMinutes,
	}
}

// Duration is the full length of the current phase.
func (t *TimerState) Duration() time.Duration {
	if t.Phase == PhaseRest {
		return time.Duration(t.RestMinutes) * time.Minute
	}
	return time.Duration(t.WorkMinutes) * time.Minute
}

// Remaining is the time left in the current phase at now, never negative.
func (t *TimerState) Remaining(now time.Time) time.Duration {
	elapsed := t.Elapsed
	if t.Status == StatusRunning {
		elapsed += now.Sub(t.RunningSince)
	}
	return max(t.Duration()-elapsed, 0)
}

// RemainingSeconds is Remaining rounded up to whole seconds, the way a
// countdown is displayed: a phase with 0.4s left still shows 1.
func (t *TimerState) RemainingSeconds(now time.Time) int {
	return int((t.Remaining(now) + time.Second - 1) / time.Second)
}

// IsActive reports whether a phase is in progress: running with time left, or
// paused. A running timer whose time has run out is no longer active.
func (t *TimerState) IsActive(now time.Time) bool {
	switch t.Status {
	case StatusPaused:
		return true
	case StatusRunning:
		return t.Remaining(now) > 0
	default:
		return false
	}
}

// Start begins a new work phase. Starting while a phase is already active is
// ignored, so a duplicate Start keeps a single active cycle (UC-05 E-4).
func (t *TimerState) Start(now time.Time) {
	if t.IsActive(now) {
		return
	}
	t.Phase = PhaseWork
	t.beginPhase(now)
}

// Pause freezes the remaining time of a running phase.
func (t *TimerState) Pause(now time.Time) error {
	if t.Status != StatusRunning || !t.IsActive(now) {
		return ErrInvalidState
	}
	t.Elapsed += now.Sub(t.RunningSince)
	t.RunningSince = time.Time{}
	t.Status = StatusPaused
	t.LastUpdated = now
	return nil
}

// Resume continues a paused phase from the time it was frozen at.
func (t *TimerState) Resume(now time.Time) error {
	if t.Status != StatusPaused {
		return ErrInvalidState
	}
	t.RunningSince = now
	t.Status = StatusRunning
	t.LastUpdated = now
	return nil
}

// Stop ends the current phase. A work phase in progress is discarded; the
// completed cycle count is kept.
func (t *TimerState) Stop(now time.Time) {
	t.Phase = PhaseWork
	t.clearProgress(now)
}

// Reset returns the timer to its initial stopped state, keeping only the
// participant's chosen work and rest lengths.
func (t *TimerState) Reset(now time.Time) {
	t.Phase = PhaseWork
	t.CurrentCycle = 0
	t.clearProgress(now)
}

// CompleteCycle ends the current phase and starts the next one. It reports
// whether the phase that ended was a work phase, which is the only kind that
// counts as a completed cycle and earns a reward.
func (t *TimerState) CompleteCycle(now time.Time) (workCompleted bool) {
	workCompleted = t.Phase == PhaseWork
	if workCompleted {
		t.CurrentCycle++
		t.Phase = PhaseRest
	} else {
		t.Phase = PhaseWork
	}
	t.beginPhase(now)
	return workCompleted
}

// SkipRest ends a rest phase early and starts the next work phase.
func (t *TimerState) SkipRest(now time.Time) error {
	if t.Phase != PhaseRest {
		return ErrInvalidState
	}
	t.Phase = PhaseWork
	t.beginPhase(now)
	return nil
}

// UpdateSetting changes the work and rest lengths. Both must be positive.
func (t *TimerState) UpdateSetting(workMinutes, restMinutes int, now time.Time) error {
	if workMinutes <= 0 || restMinutes <= 0 {
		return ErrInvalid
	}
	t.WorkMinutes = workMinutes
	t.RestMinutes = restMinutes
	t.LastUpdated = now
	return nil
}

func (t *TimerState) beginPhase(now time.Time) {
	t.Status = StatusRunning
	t.RunningSince = now
	t.Elapsed = 0
	t.LastUpdated = now
}

func (t *TimerState) clearProgress(now time.Time) {
	t.Status = StatusStopped
	t.RunningSince = time.Time{}
	t.Elapsed = 0
	t.LastUpdated = now
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
