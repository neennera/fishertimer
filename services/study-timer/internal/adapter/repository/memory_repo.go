package repository

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type memTimer struct {
	id, sessionID, userID string
	status                domain.SessionTimerStatus
	openedAt              time.Time
	finalizedAt           *time.Time
}

// InMemoryRepository keeps timers and cycles in process memory with the same
// rules as PostgresRepository. It stores and returns copies.
type InMemoryRepository struct {
	mu        sync.Mutex
	timers    []memTimer
	cycles    []domain.Cycle
	settings  map[string]domain.Settings
	processed map[string]bool
	nextID    int
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{settings: map[string]domain.Settings{}, processed: map[string]bool{}}
}

func (r *InMemoryRepository) id(prefix string) string {
	r.nextID++
	return fmt.Sprintf("%s-%d", prefix, r.nextID)
}

func (r *InMemoryRepository) findTimer(sessionID, userID string) int {
	for i, t := range r.timers {
		if t.sessionID == sessionID && t.userID == userID {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) timerByID(id string) *memTimer {
	for i := range r.timers {
		if r.timers[i].id == id {
			return &r.timers[i]
		}
	}
	return nil
}

func (r *InMemoryRepository) build(m memTimer) *domain.Timer {
	t := &domain.Timer{
		TimerID: m.id, SessionID: m.sessionID, UserID: m.userID,
		Status: m.status, OpenedAt: m.openedAt, FinalizedAt: m.finalizedAt,
		Settings: r.settingsOf(m.userID),
	}
	for i := range r.cycles {
		c := r.cycles[i]
		if c.TimerID != m.id {
			continue
		}
		if c.Active() {
			cc := c
			t.Active = &cc
			continue
		}
		if t.Last == nil || endKey(c).After(endKey(*t.Last)) {
			cc := c
			t.Last = &cc
		}
		if c.Type == domain.PhaseWork && c.Status == domain.CycleCompleted && !c.StartedAt.Before(m.openedAt) {
			t.CompletedWork++
			t.FocusSeconds += c.DurationSec
			if t.LastCompletedWork == nil || endKey(c).After(endKey(*t.LastCompletedWork)) {
				cc := c
				t.LastCompletedWork = &cc
			}
		}
	}
	return t
}

// endKey orders ended cycles; ties (same instant) keep insertion order via >=.
func endKey(c domain.Cycle) time.Time {
	if c.EndedAt != nil {
		return *c.EndedAt
	}
	return c.StartedAt
}

func (r *InMemoryRepository) settingsOf(userID string) domain.Settings {
	if s, ok := r.settings[userID]; ok {
		return s
	}
	return domain.DefaultSettings
}

func (r *InMemoryRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.Timer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := r.findTimer(sessionID, userID)
	if i < 0 {
		return nil, domain.ErrNotFound
	}
	return r.build(r.timers[i]), nil
}

func (r *InMemoryRepository) ListRoomTimers(ctx context.Context, sessionID string) ([]*domain.Timer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []*domain.Timer
	for _, t := range r.timers {
		if t.sessionID == sessionID && t.status == domain.SessionTimerOpen {
			res = append(res, r.build(t))
		}
	}
	return res, nil
}

func (r *InMemoryRepository) EnsureTimer(ctx context.Context, sessionID, userID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findTimer(sessionID, userID) < 0 {
		r.timers = append(r.timers, memTimer{id: r.id("timer"), sessionID: sessionID, userID: userID,
			status: domain.SessionTimerOpen, openedAt: now})
	}
	return nil
}

func (r *InMemoryRepository) InsertCycle(ctx context.Context, c *domain.Cycle) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.Active() {
		for _, other := range r.cycles {
			if other.TimerID == c.TimerID && other.Active() {
				return domain.ErrInvalidState
			}
		}
	}
	c.CycleID = r.id("cycle")
	// Ended cycles inserted at the same instant as an earlier one must sort
	// after it: nudge by a nanosecond.
	if c.EndedAt != nil {
		for _, other := range r.cycles {
			if other.TimerID == c.TimerID && other.EndedAt != nil && !c.EndedAt.After(*other.EndedAt) {
				at := other.EndedAt.Add(time.Nanosecond)
				c.EndedAt = &at
			}
		}
	}
	r.cycles = append(r.cycles, *c)
	return nil
}

func (r *InMemoryRepository) UpdateCycle(ctx context.Context, c *domain.Cycle, from ...domain.CycleStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.cycles {
		if r.cycles[i].CycleID == c.CycleID {
			if !slices.Contains(from, r.cycles[i].Status) {
				return domain.ErrConflict
			}
			r.cycles[i] = *c
			return nil
		}
	}
	return domain.ErrConflict
}

func (r *InMemoryRepository) GetSettings(ctx context.Context, userID string) (domain.Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settingsOf(userID), nil
}

func (r *InMemoryRepository) SaveSettings(ctx context.Context, userID string, s domain.Settings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings[userID] = s
	return nil
}

func (r *InMemoryRepository) refs(keep func(c domain.Cycle, t memTimer) bool, limit int) []domain.CycleRef {
	r.mu.Lock()
	defer r.mu.Unlock()
	var res []domain.CycleRef
	for _, c := range r.cycles {
		t := r.timerByID(c.TimerID)
		if t != nil && keep(c, *t) {
			res = append(res, domain.CycleRef{Cycle: c, SessionID: t.sessionID, UserID: t.userID})
			if len(res) == limit {
				break
			}
		}
	}
	return res
}

func (r *InMemoryRepository) DueCycles(ctx context.Context, now time.Time, limit int) ([]domain.CycleRef, error) {
	return r.refs(func(c domain.Cycle, t memTimer) bool {
		return t.status == domain.SessionTimerOpen && c.Due(now)
	}, limit), nil
}

func (r *InMemoryRepository) StalePausedCycles(ctx context.Context, cutoff time.Time, limit int) ([]domain.CycleRef, error) {
	return r.refs(func(c domain.Cycle, t memTimer) bool {
		return t.status == domain.SessionTimerOpen && c.Status == domain.CyclePaused && c.PausedAt != nil && c.PausedAt.Before(cutoff)
	}, limit), nil
}

func (r *InMemoryRepository) PendingRewards(ctx context.Context, limit int) ([]domain.CycleRef, error) {
	return r.refs(func(c domain.Cycle, t memTimer) bool {
		return c.Type == domain.PhaseWork && c.Status == domain.CycleCompleted && c.RewardStatus == domain.RewardPending
	}, limit), nil
}

func (r *InMemoryRepository) MarkRewardSent(ctx context.Context, cycleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.cycles {
		if r.cycles[i].CycleID == cycleID && r.cycles[i].RewardStatus == domain.RewardPending {
			r.cycles[i].RewardStatus = domain.RewardSent
		}
	}
	return nil
}

func (r *InMemoryRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	h := &domain.TimerHistory{UserID: userID}
	byDay := map[string]int{}
	for _, t := range r.timers {
		if t.userID != userID {
			continue
		}
		h.SessionsJoined++
		if t.openedAt.After(h.LastActive) {
			h.LastActive = t.openedAt
		}
		for _, c := range r.cycles {
			if c.TimerID != t.id || c.Type != domain.PhaseWork || c.Status != domain.CycleCompleted {
				continue
			}
			h.CyclesCompleted++
			h.TotalFocusMinutes += c.DurationSec / 60
			byDay[c.EndedAt.UTC().Format("2006-01-02")] += c.DurationSec / 60
			if c.EndedAt.After(h.LastActive) {
				h.LastActive = *c.EndedAt
			}
		}
	}
	h.DailyFocusMinutes = last30Days(byDay, time.Now().UTC())
	return h, nil
}

func (r *InMemoryRepository) OpenParticipantTimer(ctx context.Context, sessionID, userID, eventID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.processed[eventID] {
		return nil
	}
	r.processed[eventID] = true
	i := r.findTimer(sessionID, userID)
	switch {
	case i < 0:
		r.timers = append(r.timers, memTimer{id: r.id("timer"), sessionID: sessionID, userID: userID,
			status: domain.SessionTimerOpen, openedAt: now})
	case r.timers[i].status == domain.SessionTimerFinalized:
		r.timers[i].status, r.timers[i].openedAt, r.timers[i].finalizedAt = domain.SessionTimerOpen, now, nil
	}
	return nil
}

func (r *InMemoryRepository) finalize(eventID string, match func(memTimer) bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.processed[eventID] {
		return nil
	}
	r.processed[eventID] = true
	now := time.Now().UTC()
	for i := range r.timers {
		t := &r.timers[i]
		if !match(*t) || t.status != domain.SessionTimerOpen {
			continue
		}
		for j := range r.cycles {
			c := &r.cycles[j]
			if c.TimerID != t.id || !c.Active() {
				continue
			}
			if c.Due(now) {
				_ = c.Complete(now)
			} else {
				_ = c.End(domain.CycleDiscarded, now)
			}
		}
		at := now
		t.status, t.finalizedAt = domain.SessionTimerFinalized, &at
	}
	return nil
}

func (r *InMemoryRepository) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, eventType string) error {
	return r.finalize(eventID, func(t memTimer) bool { return t.sessionID == sessionID && t.userID == userID })
}

func (r *InMemoryRepository) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, eventType string) error {
	return r.finalize(eventID, func(t memTimer) bool { return t.sessionID == sessionID })
}

func (r *InMemoryRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.processed[eventID], nil
}
