package repository

import (
	"context"
	"sync"
	"github.com/neennera/fishertimer/services/account/internal/domain"
)

type InMemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]*domain.UserAccount
	profiles map[string]*domain.Profile
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		users:    make(map[string]*domain.UserAccount),
		profiles: make(map[string]*domain.Profile),
	}
}

func (r *InMemoryRepository) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Email] = u
	r.profiles[u.UserID] = &domain.Profile{
		UserID:      u.UserID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
	}
	return nil
}

func (r *InMemoryRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.users[email]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.profiles[userID]; ok {
		return p, nil
	}
	return &domain.Profile{UserID: userID, DisplayName: "Fisher", Email: "fisher@example.com"}, nil
}

func (r *InMemoryRepository) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.UserID] = p
	return nil
}
