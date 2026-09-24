package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type InMemoryRepository struct {
	mu     sync.RWMutex
	timers map[string]*domain.TimerState
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		timers: make(map[string]*domain.TimerState),
	}
}

func (r *InMemoryRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := sessionID + ":" + userID
	if t, ok := r.timers[key]; ok {
		return t, nil
	}
	return &domain.TimerState{
		SessionID:   sessionID,
		UserID:      userID,
		Status:      "STOPPED",
		Phase:       domain.PhaseWork,
		WorkMinutes: 25,
		RestMinutes: 5,
	}, nil
}

func (r *InMemoryRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := t.SessionID + ":" + t.UserID
	r.timers[key] = t
	return nil
}

func (r *InMemoryRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	return &domain.TimerHistory{
		UserID:        userID,
		TotalSessions: 12,
		TotalFocusMin: 300,
	}, nil
}
