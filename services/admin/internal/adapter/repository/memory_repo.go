package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (repo *InMemoryRepository) CreateReport(ctx context.Context, r *domain.ModerationReport) error {
	return nil
}

func (repo *InMemoryRepository) ListPending(ctx context.Context) ([]domain.ModerationReport, error) {
	return []domain.ModerationReport{{ID: "rep_1", Reason: "AFK Spam", Status: "PENDING"}}, nil
}
