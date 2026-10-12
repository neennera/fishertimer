package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// InMemoryRepository keeps rooms, participants and the outbox in process
// memory, with the same rules as PostgresRepository. One mutex stands in
// for the database transaction. It stores and returns copies, so callers can
// mutate what they get back.
type InMemoryRepository struct {
	mu           sync.Mutex
	sessions     map[string]domain.StudySession
	participants []domain.Participant // every stay, open and closed
	outbox       []domain.OutboxEvent
	nextEventID  int64
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{sessions: make(map[string]domain.StudySession)}
}

func (r *InMemoryRepository) enqueue(ev domain.OutboxEvent) {
	r.nextEventID++
	ev.ID = r.nextEventID
	r.outbox = append(r.outbox, ev)
}

// activeIndex returns the index of the user's open participation, or -1.
func (r *InMemoryRepository) activeIndex(userID string) int {
	for i, p := range r.participants {
		if p.UserID == userID && p.LeftAt == nil {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) CreateSession(ctx context.Context, s *domain.StudySession, creator *domain.Participant, joined domain.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeIndex(creator.UserID) >= 0 {
		return domain.ErrAlreadyInSession
	}
	stored := *s
	stored.ParticipantCount = 1
	stored.Status = domain.StatusActive
	r.sessions[s.ID] = stored
	r.participants = append(r.participants, *creator)
	r.enqueue(joined)
	return nil
}

func (r *InMemoryRepository) JoinSession(ctx context.Context, p *domain.Participant, joined domain.OutboxEvent) (*domain.StudySession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i := r.activeIndex(p.UserID); i >= 0 {
		if r.participants[i].SessionID != p.SessionID {
			return nil, domain.ErrAlreadyInSession
		}
		s := r.sessions[p.SessionID]
		return &s, nil
	}
	s, ok := r.sessions[p.SessionID]
	switch {
	case !ok:
		return nil, domain.ErrNotFound
	case s.Status != domain.StatusActive:
		return nil, domain.ErrSessionEnded
	case s.ParticipantCount >= s.ParticipantLimit:
		return nil, domain.ErrSessionFull
	}
	s.ParticipantCount++
	r.sessions[s.ID] = s
	r.participants = append(r.participants, *p)
	r.enqueue(joined)
	return &s, nil
}

func (r *InMemoryRepository) LeaveSession(ctx context.Context, sessionID, userID, reason string, now time.Time, left, ended domain.OutboxEvent) (*domain.LeaveResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[sessionID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	i := r.activeIndex(userID)
	if i < 0 || r.participants[i].SessionID != sessionID {
		return &domain.LeaveResult{}, nil
	}
	leftAt := now
	r.participants[i].LeftAt = &leftAt
	r.participants[i].LeaveReason = reason
	p := r.participants[i]

	s.ParticipantCount--
	r.enqueue(left)
	res := &domain.LeaveResult{Left: true, Participant: &p}
	if s.ParticipantCount == 0 && s.Status == domain.StatusActive {
		s.Status = domain.StatusEnded
		s.EndedAt = &leftAt
		s.EndReason = "EMPTY"
		r.enqueue(ended)
		res.SessionEnded = true
	}
	r.sessions[sessionID] = s
	return res, nil
}

func (r *InMemoryRepository) EndSession(ctx context.Context, sessionID, reason string, now time.Time, ended domain.OutboxEvent) (*domain.StudySession, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[sessionID]
	if !ok {
		return nil, false, domain.ErrNotFound
	}
	if s.Status == domain.StatusEnded {
		return &s, false, nil
	}
	endedAt := now
	for i, p := range r.participants {
		if p.SessionID == sessionID && p.LeftAt == nil {
			r.participants[i].LeftAt = &endedAt
			r.participants[i].LeaveReason = "SESSION_ENDED"
		}
	}
	s.Status = domain.StatusEnded
	s.EndedAt = &endedAt
	s.EndReason = reason
	s.ParticipantCount = 0
	r.sessions[sessionID] = s
	r.enqueue(ended)
	return &s, true, nil
}

func (r *InMemoryRepository) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &s, nil
}

func (r *InMemoryRepository) ListActiveSessions(ctx context.Context) ([]domain.StudySession, error) {
	return r.filterSessions(func(s domain.StudySession) bool { return s.Status == domain.StatusActive }), nil
}

func (r *InMemoryRepository) ListSessionsCreatedBefore(ctx context.Context, cutoff time.Time) ([]domain.StudySession, error) {
	return r.filterSessions(func(s domain.StudySession) bool {
		return s.Status == domain.StatusActive && s.CreatedAt.Before(cutoff)
	}), nil
}

func (r *InMemoryRepository) filterSessions(keep func(domain.StudySession) bool) []domain.StudySession {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := []domain.StudySession{}
	for _, s := range r.sessions {
		if keep(s) {
			res = append(res, s)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].CreatedAt.After(res[j].CreatedAt) })
	return res
}

func (r *InMemoryRepository) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	return r.filterParticipants(func(p domain.Participant) bool {
		return p.SessionID == sessionID && p.LeftAt == nil
	}), nil
}

func (r *InMemoryRepository) ListStaleParticipants(ctx context.Context, cutoff time.Time) ([]domain.Participant, error) {
	return r.filterParticipants(func(p domain.Participant) bool {
		return p.LeftAt == nil && p.LastSeenAt.Before(cutoff)
	}), nil
}

func (r *InMemoryRepository) filterParticipants(keep func(domain.Participant) bool) []domain.Participant {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := []domain.Participant{}
	for _, p := range r.participants {
		if keep(p) {
			res = append(res, p)
		}
	}
	return res
}

func (r *InMemoryRepository) GetActiveParticipation(ctx context.Context, userID string) (*domain.Participant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i := r.activeIndex(userID); i >= 0 {
		p := r.participants[i]
		return &p, nil
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryRepository) GetLastParticipation(ctx context.Context, sessionID, userID string) (*domain.Participant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.participants) - 1; i >= 0; i-- {
		if p := r.participants[i]; p.SessionID == sessionID && p.UserID == userID {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryRepository) TouchParticipant(ctx context.Context, sessionID, userID string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := r.activeIndex(userID)
	if i < 0 || r.participants[i].SessionID != sessionID {
		return false, nil
	}
	r.participants[i].LastSeenAt = now
	return true, nil
}

func (r *InMemoryRepository) PendingEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var pending []domain.OutboxEvent
	for _, ev := range r.outbox {
		if ev.PublishedAt == nil {
			pending = append(pending, ev)
			if len(pending) == limit {
				break
			}
		}
	}
	return pending, nil
}

func (r *InMemoryRepository) MarkEventPublished(ctx context.Context, id int64, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.outbox {
		if r.outbox[i].ID == id {
			at := now
			r.outbox[i].PublishedAt = &at
			r.outbox[i].Attempts++
		}
	}
	return nil
}

func (r *InMemoryRepository) MarkEventFailed(ctx context.Context, id int64, cause string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.outbox {
		if r.outbox[i].ID == id {
			r.outbox[i].Attempts++
		}
	}
	return nil
}
