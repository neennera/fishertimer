package domain

import "context"

// Repository is the outbound port for persisting/reading leaderboard cache data.
// In production this would be backed by Redis; in development it is in-memory.
type Repository interface {
	// GetCachedRanking returns the cached ranking for the given period, if any.
	GetCachedRanking(ctx context.Context, period string) (*CachedRanking, error)
	// SetCachedRanking stores a freshly computed ranking in the cache.
	SetCachedRanking(ctx context.Context, ranking *CachedRanking) error
}
