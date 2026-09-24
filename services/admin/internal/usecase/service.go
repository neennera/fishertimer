package usecase

import (
	"context"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

type Usecase interface {
	ViewActiveSessions(ctx context.Context) ([]domain.SessionOverview, error)
	ViewParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error)
	ViewSessionDetails(ctx context.Context, sessionID string) (*domain.SessionOverview, error)
	KickUser(ctx context.Context, sessionID, userID string) error
	EndSession(ctx context.Context, sessionID string) error
}

type service struct {
	repo          domain.Repository
	sessionClient domain.SessionClient
}

func New(repo domain.Repository, sessionClient domain.SessionClient) Usecase {
	return &service{
		repo:          repo,
		sessionClient: sessionClient,
	}
}

func (s *service) ViewActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	return s.repo.ListActiveSessions(ctx)
}

func (s *service) ViewParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	return s.repo.ListParticipants(ctx, sessionID)
}

func (s *service) ViewSessionDetails(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	return s.repo.GetSessionDetails(ctx, sessionID)
}

func (s *service) KickUser(ctx context.Context, sessionID, userID string) error {
	if s.sessionClient != nil {
		return s.sessionClient.KickParticipant(ctx, sessionID, userID)
	}
	return nil
}

func (s *service) EndSession(ctx context.Context, sessionID string) error {
	if s.sessionClient != nil {
		return s.sessionClient.CloseSession(ctx, sessionID)
	}
	return nil
}
