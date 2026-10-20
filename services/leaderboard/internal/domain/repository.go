package domain

import "context"

// Repository is the outbound port for persisting/reading leaderboard cache data.
// In production this would be backed by Redis; in development it is in-memory.
type Repository interface {
	// GetCachedRanking returns the cached ranking for the given period, if any.
	GetCachedRanking(ctx context.Context, period string) (*CachedRanking, error)
	// SetCachedRanking stores a freshly computed ranking in the cache.
	SetCachedRanking(ctx context.Context, ranking *CachedRanking) error

	// SetDisplayName remembers the name a user has chosen since their rewards
	// were recorded. Reward rows keep the name they were awarded under, so
	// without this a rename would never reach the ranking (UC-07 S-2).
	SetDisplayName(ctx context.Context, userID, displayName string) error
	// GetDisplayNames returns every remembered name, keyed by user id.
	GetDisplayNames(ctx context.Context) (map[string]string, error)
	// DeleteCachedRankings drops the cached ranking of every period, so the
	// next request recomputes it.
	DeleteCachedRankings(ctx context.Context) error
}
