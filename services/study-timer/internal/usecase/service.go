package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type Usecase interface {
	StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t := &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "RUNNING", WorkMinutes: 25, RestMinutes: 5}
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t := &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "STOPPED"}
	return t, s.repo.SaveTimer(ctx, t)
}
