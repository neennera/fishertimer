package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "STOPPED"}, nil
}

func (r *InMemoryRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	return nil
}
