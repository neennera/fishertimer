package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type InMemoryRepository struct {
	mu           sync.RWMutex
	sessions     map[string]*domain.StudySession
	participants map[string][]domain.Participant
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		sessions:     make(map[string]*domain.StudySession),
		participants: make(map[string][]domain.Participant),
	}
}

func (r *InMemoryRepository) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sess, ok := r.sessions[id]; ok {
		return sess, nil
	}
	return &domain.StudySession{ID: id, Name: "Deep Work Room", ParticipantLimit: 10, Status: "ACTIVE"}, nil
}

func (r *InMemoryRepository) CreateSession(ctx context.Context, s *domain.StudySession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *InMemoryRepository) UpdateSession(ctx context.Context, s *domain.StudySession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *InMemoryRepository) ListActiveSessions(ctx context.Context) ([]domain.StudySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.StudySession
	for _, s := range r.sessions {
		if s.Status == "ACTIVE" {
			list = append(list, *s)
		}
	}
	if len(list) == 0 {
		list = append(list, domain.StudySession{ID: "sess_demo", Name: "Focus Hall", Status: "ACTIVE", ParticipantLimit: 10})
	}
	return list, nil
}

func (r *InMemoryRepository) AddParticipant(ctx context.Context, p *domain.Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.participants[p.SessionID] = append(r.participants[p.SessionID], *p)
	return nil
}

func (r *InMemoryRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	parts := r.participants[sessionID]
	var updated []domain.Participant
	for _, p := range parts {
		if p.UserID != userID {
			updated = append(updated, p)
		}
	}
	r.participants[sessionID] = updated
	return nil
}

func (r *InMemoryRepository) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if parts, ok := r.participants[sessionID]; ok {
		return parts, nil
	}
	return []domain.Participant{{SessionID: sessionID, UserID: "usr_default"}}, nil
}
