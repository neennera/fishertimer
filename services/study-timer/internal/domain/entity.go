package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("study-timer: resource not found")
	ErrInvalid  = errors.New("study-timer: invalid input")
	// ErrInvalidState is returned when an action is not allowed in the
	// timer's current state, e.g. pausing a timer that is not running.
	ErrInvalidState = errors.New("study-timer: action not allowed in current timer state")
	// ErrTimerClosed is returned for any action on a finalized timer: the
	// user left the room, was removed, or the room ended (UC-05 E-2).
	ErrTimerClosed = errors.New("study-timer: this timer has been closed")
	// ErrNotFinished is returned when a cycle is completed before its time
	// is up; the server, not the client, decides when a cycle ends.
	ErrNotFinished = errors.New("study-timer: the cycle has time left")
	// ErrConflict is returned by the repository when a cycle changed under
	// an update (another request got there first).
	ErrConflict = errors.New("study-timer: the timer changed concurrently")
)

type TimerPhase string

const (
	PhaseWork TimerPhase = "WORK"
	PhaseRest TimerPhase = "REST"
)

// TimerStatus is the coarse status sent to clients (proto TimerStatus).
type TimerStatus string

const (
	StatusStopped TimerStatus = "STOPPED"
	StatusRunning TimerStatus = "RUNNING"
	StatusPaused  TimerStatus = "PAUSED"
)

// State is the timer's place in the UC-05 state machine.
type State string

const (
	StateReady        State = "READY"
	StateWorkRunning  State = "WORK_RUNNING"
	StateWorkPaused   State = "WORK_PAUSED"
	StateReadyForRest State = "READY_FOR_REST"
	StateRestRunning  State = "REST_RUNNING"
	StateRestPaused   State = "REST_PAUSED"
	StateFinalized    State = "FINALIZED"
)

const (
	DefaultWorkMinutes = 25
	DefaultRestMinutes = 5
)

// SessionTimerStatus tracks whether a timer in a room is open or finalized.
type SessionTimerStatus string

const (
	SessionTimerOpen      SessionTimerStatus = "OPEN"
	SessionTimerFinalized SessionTimerStatus = "FINALIZED"
)

// CycleStatus tracks the lifecycle of an individual work or rest cycle.
type CycleStatus string

const (
	CycleRunning   CycleStatus = "RUNNING"
	CyclePaused    CycleStatus = "PAUSED"
	CycleCompleted CycleStatus = "COMPLETED"
	CycleSkipped   CycleStatus = "SKIPPED"
	CycleDiscarded CycleStatus = "DISCARDED"
)

// RewardStatus tracks reward delivery for completed work cycles.
type RewardStatus string

const (
	RewardNone    RewardStatus = "NONE"
	RewardPending RewardStatus = "PENDING"
	RewardSent    RewardStatus = "SENT"
)

// Limits are the allowed cycle lengths (UC-05 E-1) and how long a cycle may
// stay paused before it is discarded (E-3).
type Limits struct {
	MinWorkMinutes  int
	MaxWorkMinutes  int
	MinRestMinutes  int
	MaxRestMinutes  int
	MaxPauseMinutes int
}

var DefaultLimits = Limits{
	MinWorkMinutes:  1,
	MaxWorkMinutes:  120,
	MinRestMinutes:  1,
	MaxRestMinutes:  60,
	MaxPauseMinutes: 15,
}

// Check validates a duration for a phase.
func (l Limits) Check(phase TimerPhase, minutes int) error {
	lo, hi := l.MinWorkMinutes, l.MaxWorkMinutes
	if phase == PhaseRest {
		lo, hi = l.MinRestMinutes, l.MaxRestMinutes
	}
	if minutes < lo || minutes > hi {
		return ErrInvalid
	}
	return nil
}

// Settings are a user's default cycle lengths (TimerSetting).
type Settings struct {
	WorkMinutes int
	RestMinutes int
}

var DefaultSettings = Settings{WorkMinutes: DefaultWorkMinutes, RestMinutes: DefaultRestMinutes}

// Cycle is one work or rest period. Its progress is derived only from
// server timestamps (UC-05 E-5): elapsed = now - started_at - paused time.
type Cycle struct {
	CycleID        string
	TimerID        string
	Type           TimerPhase
	Status         CycleStatus
	DurationSec    int
	StartedAt      time.Time
	PausedAt       *time.Time
	PausedTotalSec int
	EndedAt        *time.Time
	RewardStatus   RewardStatus
}

func (c *Cycle) Active() bool {
	return c.Status == CycleRunning || c.Status == CyclePaused
}

// Elapsed is the running time so far; a paused cycle is frozen at PausedAt.
func (c *Cycle) Elapsed(now time.Time) time.Duration {
	end := now
	if c.Status == CyclePaused && c.PausedAt != nil {
		end = *c.PausedAt
	}
	d := end.Sub(c.StartedAt) - time.Duration(c.PausedTotalSec)*time.Second
	return max(d, 0)
}

// Remaining is the time left, never negative.
func (c *Cycle) Remaining(now time.Time) time.Duration {
	return max(time.Duration(c.DurationSec)*time.Second-c.Elapsed(now), 0)
}

// Due reports whether a running cycle's time is up.
func (c *Cycle) Due(now time.Time) bool {
	return c.Status == CycleRunning && c.Remaining(now) == 0
}

// EndsAt is when a running cycle runs out.
func (c *Cycle) EndsAt() time.Time {
	return c.StartedAt.Add(time.Duration(c.DurationSec+c.PausedTotalSec) * time.Second)
}

// Pause freezes a running cycle.
func (c *Cycle) Pause(now time.Time) error {
	if c.Status != CycleRunning || c.Due(now) {
		return ErrInvalidState
	}
	at := now
	c.Status, c.PausedAt = CyclePaused, &at
	return nil
}

// Resume continues a paused cycle, adding the pause to PausedTotalSec.
func (c *Cycle) Resume(now time.Time) error {
	if c.Status != CyclePaused || c.PausedAt == nil {
		return ErrInvalidState
	}
	c.PausedTotalSec += int(now.Sub(*c.PausedAt).Round(time.Second) / time.Second)
	c.Status, c.PausedAt = CycleRunning, nil
	return nil
}

// Restart puts the active cycle back to its full length and holds it there,
// paused, until the user resumes it.
func (c *Cycle) Restart(now time.Time) error {
	if !c.Active() {
		return ErrInvalidState
	}
	at := now
	c.Status, c.StartedAt, c.PausedAt, c.PausedTotalSec = CyclePaused, now, &at, 0
	return nil
}

// Complete ends a cycle whose time is up. It ends at the moment it ran out,
// not when the completion was noticed, so focus time is exact. A completed
// work cycle owes a reward (PENDING until Reward confirms).
func (c *Cycle) Complete(now time.Time) error {
	if c.Status != CycleRunning {
		return ErrInvalidState
	}
	if !c.Due(now) {
		return ErrNotFinished
	}
	end := c.EndsAt()
	c.Status, c.EndedAt = CycleCompleted, &end
	if c.Type == PhaseWork {
		c.RewardStatus = RewardPending
	}
	return nil
}

// End closes an active cycle as discarded (stopped, abandoned, paused too
// long) or skipped (a rest period). Neither earns a reward.
func (c *Cycle) End(status CycleStatus, now time.Time) error {
	if !c.Active() {
		return ErrInvalidState
	}
	at := now
	if c.Status == CyclePaused && c.PausedAt != nil {
		at = *c.PausedAt
	}
	c.Status, c.EndedAt, c.PausedAt = status, &at, nil
	return nil
}

// Timer is one participant's timer in one room.
type Timer struct {
	TimerID     string
	SessionID   string
	UserID      string
	Status      SessionTimerStatus
	OpenedAt    time.Time
	FinalizedAt *time.Time
	Settings    Settings
	// Active is the running or paused cycle, if any.
	Active *Cycle
	// Last is the most recently ended cycle (completed, skipped or
	// discarded), if any.
	Last *Cycle
	// CompletedWork and FocusSeconds count work cycles completed since the
	// timer was (re)opened, i.e. during this stay in the room.
	CompletedWork int
	FocusSeconds  int
	// LastCompletedWork is the most recently completed work cycle.
	LastCompletedWork *Cycle
}

// NewTimer is a timer that has never been used: ready, default settings.
func NewTimer(sessionID, userID string, settings Settings, now time.Time) *Timer {
	return &Timer{SessionID: sessionID, UserID: userID, Status: SessionTimerOpen, OpenedAt: now, Settings: settings}
}

// State places the timer in the UC-05 state machine.
func (t *Timer) State() State {
	if t.Status == SessionTimerFinalized {
		return StateFinalized
	}
	if c := t.Active; c != nil {
		switch {
		case c.Type == PhaseWork && c.Status == CyclePaused:
			return StateWorkPaused
		case c.Type == PhaseWork:
			return StateWorkRunning
		case c.Status == CyclePaused:
			return StateRestPaused
		default:
			return StateRestRunning
		}
	}
	// After a completed work cycle the user picks a rest length or skips it;
	// a cycle from an earlier stay in the room does not carry over.
	if l := t.Last; l != nil && l.Type == PhaseWork && l.Status == CycleCompleted &&
		l.EndedAt != nil && !l.EndedAt.Before(t.OpenedAt) {
		return StateReadyForRest
	}
	return StateReady
}

// View is what clients see of a timer at one instant.
type View struct {
	SessionID          string
	UserID             string
	State              State
	Status             TimerStatus
	Phase              TimerPhase
	Settings           Settings
	CompletedWork      int
	FocusSeconds       int
	DurationSeconds    int
	RemainingSeconds   int
	LastUpdated        time.Time
	CycleID            string
	StartedAt          *time.Time
	PausedTotalSeconds int
	LastCompletedID    string
	LastCompletedAt    *time.Time
	Limits             Limits
}

// View renders the timer as of now.
func (t *Timer) View(now time.Time, limits Limits) View {
	v := View{
		SessionID:     t.SessionID,
		UserID:        t.UserID,
		State:         t.State(),
		Status:        StatusStopped,
		Phase:         PhaseWork,
		Settings:      t.Settings,
		CompletedWork: t.CompletedWork,
		FocusSeconds:  t.FocusSeconds,
		LastUpdated:   t.OpenedAt,
		Limits:        limits,
	}
	if t.Last != nil && t.Last.EndedAt != nil && t.Last.EndedAt.After(v.LastUpdated) {
		v.LastUpdated = *t.Last.EndedAt
	}
	if w := t.LastCompletedWork; w != nil {
		v.LastCompletedID, v.LastCompletedAt = w.CycleID, w.EndedAt
	}

	if c := t.Active; c != nil && t.Status == SessionTimerOpen {
		v.Phase = c.Type
		v.Status = StatusRunning
		if c.Status == CyclePaused {
			v.Status = StatusPaused
		}
		v.DurationSeconds = c.DurationSec
		v.RemainingSeconds = ceilSeconds(c.Remaining(now))
		v.CycleID = c.CycleID
		started := c.StartedAt
		v.StartedAt = &started
		v.PausedTotalSeconds = c.PausedTotalSec
		v.LastUpdated = c.StartedAt
		return v
	}

	minutes := t.Settings.WorkMinutes
	if v.State == StateReadyForRest {
		v.Phase, minutes = PhaseRest, t.Settings.RestMinutes
	}
	v.DurationSeconds = minutes * 60
	v.RemainingSeconds = v.DurationSeconds
	return v
}

// ceilSeconds rounds up like a countdown: 0.4s left still shows 1.
func ceilSeconds(d time.Duration) int {
	return int((d + time.Second - 1) / time.Second)
}

// CycleRef is a cycle together with whose timer it belongs to, for the
// background jobs that work across timers.
type CycleRef struct {
	Cycle
	SessionID string
	UserID    string
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
