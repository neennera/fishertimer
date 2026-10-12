package domain

import (
	"context"
)

// SessionRepository reads live room state. Study Session owns that state;
// admin_db never stores a copy of it.
type SessionRepository interface {
	ListActiveSessions(ctx context.Context) ([]SessionOverview, error)
	// GetActiveSession returns ErrNotFound when the room is not live.
	GetActiveSession(ctx context.Context, sessionID string) (*SessionOverview, error)
	// ListParticipants returns the members currently in the room.
	ListParticipants(ctx context.Context, sessionID string) ([]SessionParticipantOverview, error)
}

// AuditLogRepository persists moderation actions in admin_db.admin_logs.
type AuditLogRepository interface {
	// Record runs apply and stores entry atomically: the entry is kept only
	// when apply succeeds, and apply's effect is reported as failed when the
	// entry cannot be stored. LogID and CreatedAt are filled in on success.
	Record(ctx context.Context, entry *AdminLog, apply func(ctx context.Context) error) error
	// List returns logs newest first; an empty action means every action.
	List(ctx context.Context, action AdminAction, limit int) ([]AdminLog, error)
}
