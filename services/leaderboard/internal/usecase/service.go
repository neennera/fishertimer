package usecase

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

// Usecase is the inbound port for the Leaderboard service.
type Usecase interface {
	// ViewLeaderboard executes the full UC-08 flow:
	//   S-1 Check cache vs reward_last_update
	//   S-2 (period switch) is handled by passing a different period string
	//   S-3 Compute ranking if cache is stale
	ViewLeaderboard(ctx context.Context, period string) (*domain.CachedRanking, error)

	// RenameUser applies a display-name change from Account (profile.updated):
	// the name is remembered and the cached rankings are dropped, so the next
	// view shows the new name without waiting for another reward.
	RenameUser(ctx context.Context, userID, displayName string) error
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient

	// renameGen counts renames. A ranking computed while one landed may carry
	// the old name, so it is returned but not cached (see ViewLeaderboard).
	renameGen atomic.Uint64
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
		// Rolling 7 days window (168 hours) ensures a full 7-day week is always represented
		return now.AddDate(0, 0, -7)
	case "monthly":
		// Rolling 30 days window ensures a full month of active rewards is always represented
		return now.AddDate(0, 0, -30)
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
	gen := s.renameGen.Load()
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

	// Store in cache (ignore write errors — data is still returned to caller),
	// unless a rename landed while this was being computed: the ranking may
	// carry the old name, and caching it would keep that name on screen until
	// the TTL runs out. Skipping the write makes the next view recompute.
	if s.renameGen.Load() == gen {
		_ = s.repo.SetCachedRanking(ctx, result)
	}

	return result, nil
}

// tierScore is the points one catch is worth. The reward service sends
// score_value per catch (set by rarity tier); the rarity table is only a
// fallback for rows that arrive without one.
func tierScore(r domain.FishReward) int {
	if r.ScoreValue > 0 {
		return r.ScoreValue
	}
	switch r.Rarity {
	case "LEGENDARY":
		return 250
	case "EPIC":
		return 100
	case "RARE":
		return 50
	case "UNCOMMON":
		return 25
	default:
		return 10
	}
}

// computeRanking implements S-3: fetch all rewards, filter by period, sum the
// tier score per user, sort descending by score, break ties by earliest AwardedAt.
func (s *service) computeRanking(ctx context.Context, period string) ([]domain.RankEntry, error) {
	allRewards, err := s.rewardClient.ViewAllRewards(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUpstream, err)
	}

	// Names chosen after the rewards were recorded win over the ones on the
	// reward rows. A lookup failure only means old names show for now.
	renamed, err := s.repo.GetDisplayNames(ctx)
	if err != nil {
		log.Printf("leaderboard: could not read remembered display names: %v", err)
		renamed = nil
	}

	start := periodStart(period)

	type userAgg struct {
		displayName string
		count       int
		score       int
		earliest    time.Time // earliest AwardedAt this period (for tie-break, E-4)
	}
	agg := make(map[string]*userAgg)

	for _, r := range allRewards {
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
		a.score += tierScore(r)
		if r.AwardedAt.Before(a.earliest) {
			a.earliest = r.AwardedAt
		}
		if r.DisplayName != "" {
			a.displayName = r.DisplayName
		}
	}

	type sortable struct {
		userID string
		*userAgg
	}
	rows := make([]sortable, 0, len(agg))
	for uid, a := range agg {
		rows = append(rows, sortable{userID: uid, userAgg: a})
	}

	// S-3 sort: descending by score; ascending earliest, then id, for tie-break (E-4).
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		if !rows[i].earliest.Equal(rows[j].earliest) {
			return rows[i].earliest.Before(rows[j].earliest)
		}
		return rows[i].userID < rows[j].userID
	})

	entries := make([]domain.RankEntry, 0, len(rows))
	for rank, row := range rows {
		name := row.displayName
		if current, ok := renamed[row.userID]; ok && current != "" {
			name = current
		}
		entries = append(entries, domain.RankEntry{
			UserID:      row.userID,
			DisplayName: name,
			Rank:        rank + 1,
			RewardCount: row.count,
			Score:       row.score,
			Period:      period,
		})
	}

	return entries, nil
}

// maxDisplayNameLength is the account service's column width; the account
// service enforces a stricter limit, this only guards what is stored here.
const maxDisplayNameLength = 100

func (s *service) RenameUser(ctx context.Context, userID, displayName string) error {
	userID = strings.TrimSpace(userID)
	name := strings.TrimSpace(displayName)
	if userID == "" || name == "" || utf8.RuneCountInString(name) > maxDisplayNameLength {
		return fmt.Errorf("%w: a user id and a display name of 1-%d characters are required", domain.ErrInvalid, maxDisplayNameLength)
	}
	if err := s.repo.SetDisplayName(ctx, userID, name); err != nil {
		return err
	}
	s.renameGen.Add(1)
	// The cached rankings carry the old name. Dropping them is best effort: a
	// failure is retried with the event, and the 10 minute TTL covers the rest.
	return s.repo.DeleteCachedRankings(ctx)
}
