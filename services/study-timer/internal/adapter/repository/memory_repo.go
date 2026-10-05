package repository

import (
	"context"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// InMemoryRepository keeps timers in process memory. It stores and returns
// copies, so callers can mutate what they get back without racing each other.
type InMemoryRepository struct {
	mu              sync.RWMutex
	timers          map[string]domain.TimerState
	processedEvents map[string]bool
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		timers:          make(map[string]domain.TimerState),
		processedEvents: make(map[string]bool),
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

func (r *InMemoryRepository) GetRoomTimers(ctx context.Context, sessionID string) ([]*domain.TimerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.TimerState
	for _, t := range r.timers {
		if t.SessionID == sessionID {
			tCopy := t
			res = append(res, &tCopy)
		}
	}
	return res, nil
}

func (r *InMemoryRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timers[timerKey(t.SessionID, t.UserID)] = *t
	return nil
}

func (r *InMemoryRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processedEvents[eventID], nil
}

func (r *InMemoryRepository) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.processedEvents[eventID] {
		return nil
	}
	r.processedEvents[eventID] = true

	k := timerKey(sessionID, userID)
	if t, ok := r.timers[k]; ok {
		t.Status = domain.StatusStopped
		t.Phase = domain.PhaseWork
		t.LastUpdated = time.Now().UTC()
		r.timers[k] = t
	}
	return nil
}

func (r *InMemoryRepository) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.processedEvents[eventID] {
		return nil
	}
	r.processedEvents[eventID] = true

	now := time.Now().UTC()
	for k, t := range r.timers {
		if t.SessionID == sessionID {
			t.Status = domain.StatusStopped
			t.Phase = domain.PhaseWork
			t.LastUpdated = now
			r.timers[k] = t
		}
	}
	return nil
}

func (r *InMemoryRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	daily := make([]domain.DailyFocus, 30)
	for i := range 30 {
		date := today.AddDate(0, 0, -(29 - i)).Format("2006-01-02")
		daily[i] = domain.DailyFocus{Date: date, FocusMinutes: 0}
	}

	return &domain.TimerHistory{
		UserID:            userID,
		SessionsJoined:    12,
		CyclesCompleted:   40,
		TotalFocusMinutes: 300,
		DailyFocusMinutes: daily,
	}, nil
}
