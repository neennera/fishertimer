package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

// ── Mock Repository ───────────────────────────────────────────────────────────

type mockRepository struct {
	cache map[string]*domain.CachedRanking
}

func newMockRepo() *mockRepository {
	return &mockRepository{cache: make(map[string]*domain.CachedRanking)}
}

func (r *mockRepository) GetCachedRanking(_ context.Context, period string) (*domain.CachedRanking, error) {
	c, ok := r.cache[period]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (r *mockRepository) SetCachedRanking(_ context.Context, ranking *domain.CachedRanking) error {
	r.cache[ranking.Period] = ranking
	return nil
}

// ── Mock Reward Client ────────────────────────────────────────────────────────

type mockRewardClient struct {
	lastUpdate time.Time
	rewards    []domain.FishReward
}

func (m *mockRewardClient) ViewRewards(_ context.Context, _ string) ([]domain.FishReward, error) {
	return m.rewards, nil
}

func (m *mockRewardClient) ViewAllRewards(_ context.Context) ([]domain.FishReward, error) {
	return m.rewards, nil
}

func (m *mockRewardClient) GetLastUpdate(_ context.Context) (time.Time, error) {
	return m.lastUpdate, nil
}

// ── Seed helpers ──────────────────────────────────────────────────────────────

func seedRewards() []domain.FishReward {
	now := time.Now().UTC()
	daysAgo := func(n int) time.Time { return now.AddDate(0, 0, -n) }
	minsAgo := func(n int) time.Time { return now.Add(-time.Duration(n) * time.Minute) }
	return []domain.FishReward{
		// user1: 3 rewards all-time, none this week
		{ID: "r1", UserID: "user1", DisplayName: "Alpha", AwardedAt: daysAgo(50)},
		{ID: "r2", UserID: "user1", DisplayName: "Alpha", AwardedAt: daysAgo(40)},
		{ID: "r3", UserID: "user1", DisplayName: "Alpha", AwardedAt: daysAgo(30)},
		// user2: 5 rewards; 3 in the last 30 mins (always within any week window)
		{ID: "r4", UserID: "user2", DisplayName: "Beta", AwardedAt: daysAgo(60)},
		{ID: "r5", UserID: "user2", DisplayName: "Beta", AwardedAt: daysAgo(45)},
		{ID: "r6", UserID: "user2", DisplayName: "Beta", AwardedAt: minsAgo(30)},
		{ID: "r7", UserID: "user2", DisplayName: "Beta", AwardedAt: minsAgo(20)},
		{ID: "r8", UserID: "user2", DisplayName: "Beta", AwardedAt: minsAgo(10)},
	}
}



// ── Tests ─────────────────────────────────────────────────────────────────────

func TestViewLeaderboard_AllTime(t *testing.T) {
	repo := newMockRepo()
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: seedRewards()}
	svc := usecase.New(repo, rc)

	result, err := svc.ViewLeaderboard(context.Background(), "all-time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rankings) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result.Rankings))
	}
	// user2 has 5 all-time rewards — should rank #1
	if result.Rankings[0].UserID != "user2" {
		t.Errorf("expected user2 at rank 1, got %s", result.Rankings[0].UserID)
	}
	if result.Rankings[0].RewardCount != 5 {
		t.Errorf("expected count 5, got %d", result.Rankings[0].RewardCount)
	}
}

func TestViewLeaderboard_Weekly(t *testing.T) {
	repo := newMockRepo()
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: seedRewards()}
	svc := usecase.New(repo, rc)

	result, err := svc.ViewLeaderboard(context.Background(), "weekly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only user2 has rewards this week
	if len(result.Rankings) != 1 {
		t.Fatalf("expected 1 user with weekly rewards, got %d", len(result.Rankings))
	}
	if result.Rankings[0].UserID != "user2" {
		t.Errorf("expected user2 as weekly leader, got %s", result.Rankings[0].UserID)
	}
	if result.Rankings[0].RewardCount != 3 {
		t.Errorf("expected weekly count 3, got %d", result.Rankings[0].RewardCount)
	}
}

func TestViewLeaderboard_CacheHit(t *testing.T) {
	repo := newMockRepo()
	// Pre-seed a cached ranking with an old timestamp
	fetchedAt := time.Now().UTC().Add(-1 * time.Hour)
	repo.SetCachedRanking(context.Background(), &domain.CachedRanking{
		Period:           "all-time",
		Rankings:         []domain.RankEntry{{UserID: "cached_user", Rank: 1, RewardCount: 99}},
		LeaderboardFetch: fetchedAt,
	})

	// reward_last_update is older than leaderboard_last_fetch → cache should be used
	rc := &mockRewardClient{
		lastUpdate: fetchedAt.Add(-30 * time.Minute), // reward updated before our cache
		rewards:    seedRewards(),
	}
	svc := usecase.New(repo, rc)

	result, err := svc.ViewLeaderboard(context.Background(), "all-time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected cache hit (result.Cached == true)")
	}
	if result.Rankings[0].UserID != "cached_user" {
		t.Errorf("expected cached_user from cache, got %s", result.Rankings[0].UserID)
	}
}

func TestViewLeaderboard_CacheMiss_Recomputes(t *testing.T) {
	repo := newMockRepo()
	// Pre-seed a cached ranking
	fetchedAt := time.Now().UTC().Add(-1 * time.Hour)
	repo.SetCachedRanking(context.Background(), &domain.CachedRanking{
		Period:           "all-time",
		Rankings:         []domain.RankEntry{{UserID: "stale_user", Rank: 1, RewardCount: 1}},
		LeaderboardFetch: fetchedAt,
	})

	// reward_last_update is NEWER than cache → should recompute
	rewardUpdate := time.Now().UTC()
	rc := &mockRewardClient{
		lastUpdate: rewardUpdate,
		rewards:    seedRewards(),
	}
	svc := usecase.New(repo, rc)

	result, err := svc.ViewLeaderboard(context.Background(), "all-time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cached {
		t.Error("expected cache miss (result.Cached == false)")
	}
	if result.Rankings[0].UserID != "user2" {
		t.Errorf("expected recomputed rank1=user2, got %s", result.Rankings[0].UserID)
	}
	// S-1 spec: leaderboard_last_fetch must equal reward_last_update
	if !result.LeaderboardFetch.Equal(rewardUpdate) {
		t.Errorf("expected leaderboard_last_fetch == reward_last_update (%v), got %v",
			rewardUpdate, result.LeaderboardFetch)
	}
}

func TestViewLeaderboard_EmptyState(t *testing.T) {
	repo := newMockRepo()
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: nil}
	svc := usecase.New(repo, rc)

	result, err := svc.ViewLeaderboard(context.Background(), "weekly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// E-1: empty period → empty rankings, not an error
	if len(result.Rankings) != 0 {
		t.Errorf("expected 0 rankings for empty week, got %d", len(result.Rankings))
	}
}
