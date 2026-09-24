package domain

import (
	"context"
)

type Repository interface {
	GetSession(ctx context.Context, id string) (*StudySession, error)
	CreateSession(ctx context.Context, s *StudySession) error
	UpdateSession(ctx context.Context, s *StudySession) error
	ListActiveSessions(ctx context.Context) ([]StudySession, error)
	AddParticipant(ctx context.Context, p *Participant) error
	RemoveParticipant(ctx context.Context, sessionID, userID string) error
	GetParticipants(ctx context.Context, sessionID string) ([]Participant, error)
}
