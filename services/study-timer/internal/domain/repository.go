package domain

import (
	"context"
)

type Repository interface {
	// GetTimer returns ErrNotFound when the participant has no timer yet.
	GetTimer(ctx context.Context, sessionID, userID string) (*TimerState, error)
	// GetRoomTimers returns all participant timers in a room session.
	GetRoomTimers(ctx context.Context, sessionID string) ([]*TimerState, error)
	SaveTimer(ctx context.Context, t *TimerState) error
	GetHistory(ctx context.Context, userID string) (*TimerHistory, error)

	// Phase 2: Idempotent Event-driven finalization
	FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error
	FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
}
