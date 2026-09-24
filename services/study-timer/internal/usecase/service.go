package usecase

import (
	"context"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type Usecase interface {
	StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	PauseTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	ResetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	SkipRest(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.TimerState, error)
	TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error)
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient
}

func New(repo domain.Repository, rewardClient domain.RewardClient) Usecase {
	return &service{
		repo:         repo,
		rewardClient: rewardClient,
	}
}

func (s *service) StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		t = &domain.TimerState{
			SessionID:   sessionID,
			UserID:      userID,
			WorkMinutes: 25,
			RestMinutes: 5,
		}
	}
	t.Status = "RUNNING"
	t.Phase = domain.PhaseWork
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) PauseTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	t.Status = "PAUSED"
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	t.Status = "RUNNING"
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	t.Status = "STOPPED"
	t.Phase = domain.PhaseWork
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) ResetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	t.Status = "STOPPED"
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	// 1. Call Reward Service AwardReward() collaborator
	if s.rewardClient != nil {
		_ = s.rewardClient.AwardReward(ctx, userID, "CompleteCycle")
	}

	// 2. Increment cycle count and switch phase
	t.CurrentCycle++
	if t.Phase == domain.PhaseWork {
		t.Phase = domain.PhaseRest
	} else {
		t.Phase = domain.PhaseWork
	}
	t.Status = "RUNNING"
	t.LastUpdated = time.Now().UTC()

	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) SkipRest(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	t.Phase = domain.PhaseWork
	t.Status = "RUNNING"
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.TimerState, error) {
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if err != nil {
		t = &domain.TimerState{
			SessionID: sessionID,
			UserID:    userID,
		}
	}
	t.WorkMinutes = workMin
	t.RestMinutes = restMin
	t.LastUpdated = time.Now().UTC()
	return t, s.repo.SaveTimer(ctx, t)
}

func (s *service) TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	return s.repo.GetHistory(ctx, userID)
}
