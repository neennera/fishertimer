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
	names map[string]string
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{
		cache: make(map[string]*domain.CachedRanking),
		names: make(map[string]string),
	}
}

func (r *InMemoryRepository) SetDisplayName(ctx context.Context, userID, displayName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.names[userID] = displayName
	return nil
}

func (r *InMemoryRepository) GetDisplayNames(ctx context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.names))
	for k, v := range r.names {
		out[k] = v
	}
	return out, nil
}

func (r *InMemoryRepository) DeleteCachedRankings(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]*domain.CachedRanking)
	return nil
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
