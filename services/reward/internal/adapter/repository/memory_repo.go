package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (repo *InMemoryRepository) Award(ctx context.Context, r *domain.FishReward) error {
	return nil
}

func (repo *InMemoryRepository) ListByUser(ctx context.Context, userID string) ([]domain.FishReward, error) {
	return []domain.FishReward{{ID: "fish_1", UserID: userID, Species: "Golden Salmon", Rarity: "RARE"}}, nil
}
