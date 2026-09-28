package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

// Usecase is the inbound port for the Leaderboard service.
type Usecase interface {
	// ViewLeaderboard executes the full UC-08 flow:
	//   S-1 Check cache vs reward_last_update
	//   S-2 (period switch) is handled by passing a different period string
	//   S-3 Compute ranking if cache is stale
	ViewLeaderboard(ctx context.Context, period string) (*domain.CachedRanking, error)
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient
}

func New(repo domain.Repository, rewardClient domain.RewardClient) Usecase {
	return &service{repo: repo, rewardClient: rewardClient}
}

// periodStart returns the UTC start of the window for the requested period.
// "all-time" returns the zero time so all rewards qualify.
func periodStart(period string) time.Time {
	now := time.Now().UTC()
	switch period {
	case "weekly":
		// ISO week starts on Monday.
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // treat Sunday as day 7 so Monday is always day 1
		}
		return time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	case "monthly":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default: // "all-time"
		return time.Time{}
	}
}

// ViewLeaderboard implements UC-08 including the S-1 cache-validation subflow.
//
// S-1: Compare reward_last_update with leaderboard_last_fetch.
//   - Equal  → return cached ranking (cache hit).
//   - Differ → recompute (S-3), persist, update leaderboard_last_fetch.
//
// S-3: Sum reward_count per user within the period, order descending.
//
//	Tie-break: user who reached that total first (earliest awarded_at).
func (s *service) ViewLeaderboard(ctx context.Context, period string) (*domain.CachedRanking, error) {
	// Normalise period value.
	switch period {
	case "weekly", "monthly", "all-time":
	default:
		period = "all-time"
	}

	// S-1 — fetch reward_last_update from Reward service.
	rewardLastUpdate, err := s.rewardClient.GetLastUpdate(ctx)
	if err != nil {
		// Degrade gracefully: if the reward service is unreachable, try the cache.
		rewardLastUpdate = time.Time{}
	}

	// S-1 — check cache.
	cached, _ := s.repo.GetCachedRanking(ctx, period)
	if cached != nil && !rewardLastUpdate.IsZero() &&
		!cached.LeaderboardFetch.Before(rewardLastUpdate) {
		// Cache is fresh — return it (S-1 cache hit).
		cached.Cached = true
		return cached, nil
	}

	// S-3 — recompute ranking.
	rankings, err := s.computeRanking(ctx, period)
	if err != nil {
		// If recomputation fails but we have a stale cache, return it.
		if cached != nil {
			cached.Cached = true
			return cached, nil
		}
		return nil, err
	}

	// S-1 spec: leaderboard_last_fetch = reward_last_update (not time.Now).
	// This ensures future comparisons hit the cache when reward_last_update is unchanged.
	fetchTime := rewardLastUpdate
	if fetchTime.IsZero() {
		fetchTime = time.Now().UTC()
	}

	result := &domain.CachedRanking{
		Rankings:         rankings,
		Period:           period,
		LeaderboardFetch: fetchTime,
		Cached:           false,
	}

	// Store in cache (ignore write errors — data is still returned to caller).
	_ = s.repo.SetCachedRanking(ctx, result)

	return result, nil
}

// computeRanking implements S-3: fetch all rewards, filter by period, aggregate per user,
// sort descending by count, break ties by earliest AwardedAt.
func (s *service) computeRanking(ctx context.Context, period string) ([]domain.RankEntry, error) {
	allRewards, err := s.rewardClient.ViewAllRewards(ctx)
	if err != nil {
		return nil, err
	}

	start := periodStart(period)

	// Aggregate: count rewards per user within the period window.
	type userAgg struct {
		displayName string
		count       int
		earliest    time.Time // earliest AwardedAt this period (for tie-break, E-4)
	}
	agg := make(map[string]*userAgg)

	for _, r := range allRewards {
		// E-4 tie-break: use AwardedAt; skip zero-time rewards to be safe.
		if !start.IsZero() && r.AwardedAt.Before(start) {
			continue
		}
		if _, ok := agg[r.UserID]; !ok {
			agg[r.UserID] = &userAgg{
				displayName: r.DisplayName,
				earliest:    r.AwardedAt,
			}
		}
		a := agg[r.UserID]
		a.count++
		if r.AwardedAt.Before(a.earliest) {
			a.earliest = r.AwardedAt
		}
		// Keep display name from most recent entry if seed has it populated.
		if r.DisplayName != "" {
			a.displayName = r.DisplayName
		}
	}

	// Build sortable slice.
	type sortable struct {
		userID      string
		displayName string
		count       int
		earliest    time.Time
	}
	var rows []sortable
	for uid, a := range agg {
		rows = append(rows, sortable{
			userID:      uid,
			displayName: a.displayName,
			count:       a.count,
			earliest:    a.earliest,
		})
	}

	// S-3 sort: descending by count; ascending earliest for tie-break (E-4).
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count != rows[j].count {
			return rows[i].count > rows[j].count
		}
		return rows[i].earliest.Before(rows[j].earliest)
	})

	entries := make([]domain.RankEntry, 0, len(rows))
	for rank, row := range rows {
		entries = append(entries, domain.RankEntry{
			UserID:      row.userID,
			DisplayName: row.displayName,
			Rank:        rank + 1,
			RewardCount: row.count,
			Period:      period,
		})
	}

	return entries, nil
}
