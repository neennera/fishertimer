package domain

import (
	"context"
	"time"
)

// Repository persists timers, cycles, settings and processed events in
// timer_db (timers, cycles, timer_settings, processed_events).
type Repository interface {
	// GetTimer loads a timer with its active cycle, last cycle and the work
	// completed during the current stay. ErrNotFound if the user has no
	// timer in this room yet.
	GetTimer(ctx context.Context, sessionID, userID string) (*Timer, error)
	// ListRoomTimers returns the room's OPEN timers.
	ListRoomTimers(ctx context.Context, sessionID string) ([]*Timer, error)
	// EnsureTimer creates an OPEN timer if the user has none in this room
	// yet. An existing timer, open or finalized, is left as it is.
	EnsureTimer(ctx context.Context, sessionID, userID string, now time.Time) error

	// InsertCycle stores a new cycle. ErrInvalidState if the timer already
	// has an active cycle (unique index, UC-05 E-4).
	InsertCycle(ctx context.Context, c *Cycle) error
	// UpdateCycle saves c only if its stored status is still one of from;
	// otherwise ErrConflict.
	UpdateCycle(ctx context.Context, c *Cycle, from ...CycleStatus) error

	GetSettings(ctx context.Context, userID string) (Settings, error)
	SaveSettings(ctx context.Context, userID string, s Settings) error

	// DueCycles returns running cycles of open timers whose time is up.
	DueCycles(ctx context.Context, now time.Time, limit int) ([]CycleRef, error)
	// StalePausedCycles returns cycles paused since before cutoff.
	StalePausedCycles(ctx context.Context, cutoff time.Time, limit int) ([]CycleRef, error)
	// PendingRewards returns completed work cycles whose reward has not been
	// confirmed by the Reward service yet.
	PendingRewards(ctx context.Context, limit int) ([]CycleRef, error)
	MarkRewardSent(ctx context.Context, cycleID string) error

	GetHistory(ctx context.Context, userID string) (*TimerHistory, error)

	// Event-driven, idempotent by event_id (processed_events).
	// OpenParticipantTimer (re)opens the user's timer when they join.
	OpenParticipantTimer(ctx context.Context, sessionID, userID, eventID string, now time.Time) error
	// Finalize* discard active cycles and mark timers FINALIZED.
	FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error
	FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
}
