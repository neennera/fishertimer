package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type Usecase interface {
	StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	GetRoomTimers(ctx context.Context, sessionID string) ([]*domain.TimerState, error)
	PauseTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	ResetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	SkipRest(ctx context.Context, sessionID, userID string) (*domain.TimerState, error)
	UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.TimerState, error)
	TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error)

	// Phase 2: RabbitMQ Event Consumers
	FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, reason string) error
	FinalizeSessionTimers(ctx context.Context, sessionID, eventID, reason string) error
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient
	now          func() time.Time
}

// Option customises the usecase service.
type Option func(*service)

// WithClock replaces the wall clock, so tests can control elapsed time.
func WithClock(now func() time.Time) Option {
	return func(s *service) { s.now = now }
}

func New(repo domain.Repository, rewardClient domain.RewardClient, opts ...Option) Usecase {
	s := &service{
		repo:         repo,
		rewardClient: rewardClient,
		now:          func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// loadTimer returns the participant's timer, or a new stopped one if they
// have never used the timer in this room.
func (s *service) loadTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	if sessionID == "" || userID == "" {
		return nil, domain.ErrInvalid
	}
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.NewTimer(sessionID, userID), nil
	}
	return t, err
}

// update loads the timer, applies change to it and saves the result.
func (s *service) update(ctx context.Context, sessionID, userID string, change func(t *domain.TimerState, now time.Time) error) (*domain.TimerState, error) {
	t, err := s.loadTimer(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if err := change(t, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.SaveTimer(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) StartTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, func(t *domain.TimerState, now time.Time) error {
		t.Start(now)
		return nil
	})
}

func (s *service) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.loadTimer(ctx, sessionID, userID)
}

func (s *service) PauseTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, (*domain.TimerState).Pause)
}

func (s *service) ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, (*domain.TimerState).Resume)
}

func (s *service) StopTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, func(t *domain.TimerState, now time.Time) error {
		t.Stop(now)
		return nil
	})
}

func (s *service) ResetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, func(t *domain.TimerState, now time.Time) error {
		t.Reset(now)
		return nil
	})
}

func (s *service) CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	var workCompleted bool
	t, err := s.update(ctx, sessionID, userID, func(t *domain.TimerState, now time.Time) error {
		workCompleted = t.CompleteCycle(now)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Only a completed work cycle earns a reward (UC-05 S-2); rest earns nothing.
	if workCompleted && s.rewardClient != nil {
		_ = s.rewardClient.AwardReward(ctx, userID, "CompleteCycle")
	}
	return t, nil
}

func (s *service) SkipRest(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, (*domain.TimerState).SkipRest)
}

func (s *service) UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.TimerState, error) {
	return s.update(ctx, sessionID, userID, func(t *domain.TimerState, now time.Time) error {
		return t.UpdateSetting(workMin, restMin, now)
	})
}

func (s *service) TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	return s.repo.GetHistory(ctx, userID)
}

func (s *service) GetRoomTimers(ctx context.Context, sessionID string) ([]*domain.TimerState, error) {
	if sessionID == "" {
		return nil, domain.ErrInvalid
	}
	return s.repo.GetRoomTimers(ctx, sessionID)
}

func (s *service) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, reason string) error {
	if sessionID == "" || userID == "" || eventID == "" {
		return domain.ErrInvalid
	}
	return s.repo.FinalizeParticipantTimer(ctx, sessionID, userID, eventID, "session.participant.left")
}

func (s *service) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, reason string) error {
	if sessionID == "" || eventID == "" {
		return domain.ErrInvalid
	}
	return s.repo.FinalizeSessionTimers(ctx, sessionID, eventID, "session.ended")
}
