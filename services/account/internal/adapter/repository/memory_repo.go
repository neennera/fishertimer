package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/account/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return &domain.Profile{UserID: userID, DisplayName: "Fisher"}, nil
}

func (r *InMemoryRepository) UpdateBanStatus(ctx context.Context, userID string, banned bool) error {
	return nil
}
