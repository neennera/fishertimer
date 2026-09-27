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

func (repo *InMemoryRepository) Award(ctx context.Context, r *domain.UnlockedReward) error {
	return nil
}

func (repo *InMemoryRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return []domain.UnlockedReward{{ItemID: "test-item", UserID: userID, ItemName: "Golden Salmon", Category: domain.CategoryFish}}, nil
}
