package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

const (
	// TTL for cached rankings. After this time Redis evicts the key automatically.
	// On the next request the service recomputes from Reward.
	cacheTTL = 10 * time.Minute

	keyPrefix = "leaderboard:ranking:"

	// namesKey is a hash of user id to the display name chosen after the
	// rewards were recorded. No TTL: it must outlive the ranking cache.
	namesKey = "leaderboard:names"
)

// periods are the cached rankings, one key each.
var periods = []string{"weekly", "monthly", "all-time"}

// RedisRepository implements domain.Repository backed by Redis.
//
// Each period's ranking is stored as a JSON-encoded CachedRanking under the key
// "leaderboard:ranking:<period>" with a TTL as a safety net (the primary
// invalidation mechanism is still the reward_last_update comparison in the usecase).
type RedisRepository struct {
	client *redis.Client
}

// NewRedis creates a RedisRepository from a redis.Client.
// Call this from main.go after verifying the connection with Ping.
func NewRedis(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// NewRedisFromURL parses a Redis URL and returns a ready-to-use repository.
// Returns an error if the URL is invalid or the server is unreachable.
func NewRedisFromURL(ctx context.Context, redisURL string) (*RedisRepository, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return NewRedis(client), nil
}

// GetCachedRanking fetches the JSON-encoded CachedRanking from Redis.
// Returns (nil, nil) when the key does not exist (cache miss).
func (r *RedisRepository) GetCachedRanking(ctx context.Context, period string) (*domain.CachedRanking, error) {
	key := keyPrefix + period
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // cache miss — not an error
		}
		return nil, err
	}

	var ranking domain.CachedRanking
	if err := json.Unmarshal(data, &ranking); err != nil {
		return nil, err
	}
	return &ranking, nil
}

// SetCachedRanking persists the ranking as JSON in Redis with a TTL.
func (r *RedisRepository) SetCachedRanking(ctx context.Context, ranking *domain.CachedRanking) error {
	key := keyPrefix + ranking.Period
	data, err := json.Marshal(ranking)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, cacheTTL).Err()
}

// SetDisplayName remembers a renamed user's current name.
func (r *RedisRepository) SetDisplayName(ctx context.Context, userID, displayName string) error {
	return r.client.HSet(ctx, namesKey, userID, displayName).Err()
}

// GetDisplayNames returns every remembered name, keyed by user id.
func (r *RedisRepository) GetDisplayNames(ctx context.Context) (map[string]string, error) {
	names, err := r.client.HGetAll(ctx, namesKey).Result()
	if err != nil {
		return nil, err
	}
	return names, nil
}

// DeleteCachedRankings drops every period's cached ranking.
func (r *RedisRepository) DeleteCachedRankings(ctx context.Context) error {
	keys := make([]string, len(periods))
	for i, p := range periods {
		keys[i] = keyPrefix + p
	}
	return r.client.Del(ctx, keys...).Err()
}

// Close releases the underlying Redis connection pool.
func (r *RedisRepository) Close() error {
	return r.client.Close()
}
