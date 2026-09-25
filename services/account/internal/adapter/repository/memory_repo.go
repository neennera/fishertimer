package repository

import (
	"context"
	"strings"
	"sync"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// InMemoryRepository is the unit-test double for domain.Repository
// (docs/skill-set/4-database-and-repository). It is NOT wired into the running
// service: account_db is required at startup, because a store that loses every
// account on restart is worse than a service that refuses to start.
type InMemoryRepository struct {
	mu      sync.RWMutex
	byID    map[string]*domain.UserAccount
	byEmail map[string]string
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		byID:    make(map[string]*domain.UserAccount),
		byEmail: make(map[string]string),
	}
}

func (r *InMemoryRepository) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	if u == nil || u.UserID == "" {
		return domain.ErrInvalid
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[u.UserID] = clone(u)
	r.byEmail[strings.ToLower(u.Email)] = u.UserID
	return nil
}

func (r *InMemoryRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return clone(r.byID[id]), nil
}

func (r *InMemoryRepository) GetUserByID(ctx context.Context, userID string) (*domain.UserAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byID[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return clone(u), nil
}

func clone(u *domain.UserAccount) *domain.UserAccount {
	if u == nil {
		return nil
	}
	copied := *u
	return &copied
}
