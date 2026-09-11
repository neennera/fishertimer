package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/auth/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) FindByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	return nil, domain.ErrNotFound
}

func (r *InMemoryRepository) Create(ctx context.Context, user *domain.UserAccount) error {
	return nil
}
