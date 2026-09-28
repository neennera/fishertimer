package repository

import (
	"context"
	"sync"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

// InMemoryRepository is the development/test implementation of domain.Repository.
// In production this would delegate to Redis using the REDIS_URL from config.
//
// It stores one CachedRanking per period key ("weekly", "monthly", "all-time").
type InMemoryRepository struct {
	mu    sync.RWMutex
	cache map[string]*domain.CachedRanking
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		cache: make(map[string]*domain.CachedRanking),
	}
}

// GetCachedRanking returns the cached ranking for the given period, or nil if
// no cache entry exists yet (first request or evicted).
func (r *InMemoryRepository) GetCachedRanking(ctx context.Context, period string) (*domain.CachedRanking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cache[period]
	if !ok {
		return nil, nil
	}
	// Return a shallow copy so callers cannot mutate the stored entry.
	copy := *c
	return &copy, nil
}

// SetCachedRanking stores the freshly computed ranking for the given period.
func (r *InMemoryRepository) SetCachedRanking(ctx context.Context, ranking *domain.CachedRanking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *ranking
	r.cache[ranking.Period] = &copy
	return nil
}
