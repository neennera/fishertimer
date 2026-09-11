package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (repo *InMemoryRepository) GetRankings(ctx context.Context, period string) ([]domain.RankEntry, error) {
	return []domain.RankEntry{{UserID: "u1", DisplayName: "TopAngler", Rank: 1, RewardCount: 50, FocusMinutes: 1200, Period: period}}, nil
}

func (repo *InMemoryRepository) UpsertScore(ctx context.Context, entry *domain.RankEntry) error {
	return nil
}
