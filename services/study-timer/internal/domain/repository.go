package domain

import (
	"context"
)

type Repository interface {
	// GetTimer returns ErrNotFound when the participant has no timer yet.
	GetTimer(ctx context.Context, sessionID, userID string) (*TimerState, error)
	SaveTimer(ctx context.Context, t *TimerState) error
	GetHistory(ctx context.Context, userID string) (*TimerHistory, error)
}
