package repository

import (
	"context"
	"sync"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// InMemoryRepository keeps timers in process memory. It stores and returns
// copies, so callers can mutate what they get back without racing each other.
type InMemoryRepository struct {
	mu     sync.RWMutex
	timers map[string]domain.TimerState
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		timers: make(map[string]domain.TimerState),
	}
}

func timerKey(sessionID, userID string) string {
	return sessionID + ":" + userID
}

func (r *InMemoryRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.timers[timerKey(sessionID, userID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &t, nil
}

func (r *InMemoryRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timers[timerKey(t.SessionID, t.UserID)] = *t
	return nil
}

func (r *InMemoryRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	return &domain.TimerHistory{
		UserID:        userID,
		TotalSessions: 12,
		TotalFocusMin: 300,
	}, nil
}
