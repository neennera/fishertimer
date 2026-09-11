package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
	return &domain.StudySession{ID: id, Name: "Deep Work Room", ParticipantLimit: 10, Status: "ACTIVE"}, nil
}

func (r *InMemoryRepository) CreateSession(ctx context.Context, s *domain.StudySession) error {
	return nil
}
