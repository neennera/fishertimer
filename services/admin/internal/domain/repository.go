package domain

import (
	"context"
)

type Repository interface {
	ListActiveSessions(ctx context.Context) ([]SessionOverview, error)
	ListParticipants(ctx context.Context, sessionID string) ([]SessionParticipantOverview, error)
	GetSessionDetails(ctx context.Context, sessionID string) (*SessionOverview, error)
}
