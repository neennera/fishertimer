package repository

import (
	"context"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

// Seed timestamps — all relative to "now" so period filters are always live.
//
//	user1: 12 rewards over 3 months  → all-time #1
//	user2:  8 rewards (2 this week)  → weekly #2
//	user3:  5 rewards (5 this week)  → weekly #1, monthly top-5
//	user4:  6 rewards (3 this month) → monthly contender
//	user5: 10 rewards (8 this month) → monthly #1
var seedRewards = func() []domain.FishReward {
	now := time.Now().UTC()

	// helper: returns a time N days ago
	daysAgo := func(n int) time.Time {
		return now.AddDate(0, 0, -n)
	}

	return []domain.FishReward{
		// ── user1: 12 rewards spread over ~90 days (all-time king) ──
		{ID: "r-u1-01", UserID: "user1", DisplayName: "TideAngler", Species: "Legendary Koi", Rarity: "LEGENDARY", AwardedAt: daysAgo(90)},
		{ID: "r-u1-02", UserID: "user1", DisplayName: "TideAngler", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(80)},
		{ID: "r-u1-03", UserID: "user1", DisplayName: "TideAngler", Species: "Rainbow Trout", Rarity: "RARE", AwardedAt: daysAgo(70)},
		{ID: "r-u1-04", UserID: "user1", DisplayName: "TideAngler", Species: "Golden Carp", Rarity: "RARE", AwardedAt: daysAgo(60)},
		{ID: "r-u1-05", UserID: "user1", DisplayName: "TideAngler", Species: "Phantom Eel", Rarity: "EPIC", AwardedAt: daysAgo(50)},
		{ID: "r-u1-06", UserID: "user1", DisplayName: "TideAngler", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(42)},
		{ID: "r-u1-07", UserID: "user1", DisplayName: "TideAngler", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(35)},
		{ID: "r-u1-08", UserID: "user1", DisplayName: "TideAngler", Species: "River Perch", Rarity: "COMMON", AwardedAt: daysAgo(28)},
		{ID: "r-u1-09", UserID: "user1", DisplayName: "TideAngler", Species: "Blue Tuna", Rarity: "EPIC", AwardedAt: daysAgo(21)},
		{ID: "r-u1-10", UserID: "user1", DisplayName: "TideAngler", Species: "Glowing Jellyfish", Rarity: "RARE", AwardedAt: daysAgo(14)},
		{ID: "r-u1-11", UserID: "user1", DisplayName: "TideAngler", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(10)},
		{ID: "r-u1-12", UserID: "user1", DisplayName: "TideAngler", Species: "River Perch", Rarity: "COMMON", AwardedAt: daysAgo(9)},

		// ── user2: 8 rewards; 2 this week ──
		{ID: "r-u2-01", UserID: "user2", DisplayName: "CastMaster", Species: "Golden Salmon", Rarity: "RARE", AwardedAt: daysAgo(75)},
		{ID: "r-u2-02", UserID: "user2", DisplayName: "CastMaster", Species: "Phantom Eel", Rarity: "EPIC", AwardedAt: daysAgo(55)},
		{ID: "r-u2-03", UserID: "user2", DisplayName: "CastMaster", Species: "River Perch", Rarity: "COMMON", AwardedAt: daysAgo(40)},
		{ID: "r-u2-04", UserID: "user2", DisplayName: "CastMaster", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(30)},
		{ID: "r-u2-05", UserID: "user2", DisplayName: "CastMaster", Species: "Rainbow Trout", Rarity: "RARE", AwardedAt: daysAgo(22)},
		{ID: "r-u2-06", UserID: "user2", DisplayName: "CastMaster", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(15)},
		{ID: "r-u2-07", UserID: "user2", DisplayName: "CastMaster", Species: "Blue Tuna", Rarity: "EPIC", AwardedAt: daysAgo(3)},  // this week
		{ID: "r-u2-08", UserID: "user2", DisplayName: "CastMaster", Species: "River Perch", Rarity: "COMMON", AwardedAt: daysAgo(1)}, // this week

		// ── user3: 5 rewards ALL this week → weekly #1 ──
		{ID: "r-u3-01", UserID: "user3", DisplayName: "LureQueen", Species: "Glowing Jellyfish", Rarity: "RARE", AwardedAt: daysAgo(6)},
		{ID: "r-u3-02", UserID: "user3", DisplayName: "LureQueen", Species: "Rainbow Trout", Rarity: "RARE", AwardedAt: daysAgo(5)},
		{ID: "r-u3-03", UserID: "user3", DisplayName: "LureQueen", Species: "Golden Carp", Rarity: "RARE", AwardedAt: daysAgo(4)},
		{ID: "r-u3-04", UserID: "user3", DisplayName: "LureQueen", Species: "Blue Tuna", Rarity: "EPIC", AwardedAt: daysAgo(2)},
		{ID: "r-u3-05", UserID: "user3", DisplayName: "LureQueen", Species: "Phantom Eel", Rarity: "EPIC", AwardedAt: daysAgo(1)},

		// ── user4: 6 rewards; 3 this month ──
		{ID: "r-u4-01", UserID: "user4", DisplayName: "DeepDiver", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(85)},
		{ID: "r-u4-02", UserID: "user4", DisplayName: "DeepDiver", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(65)},
		{ID: "r-u4-03", UserID: "user4", DisplayName: "DeepDiver", Species: "River Perch", Rarity: "COMMON", AwardedAt: daysAgo(45)},
		{ID: "r-u4-04", UserID: "user4", DisplayName: "DeepDiver", Species: "Golden Salmon", Rarity: "RARE", AwardedAt: daysAgo(20)}, // this month
		{ID: "r-u4-05", UserID: "user4", DisplayName: "DeepDiver", Species: "Rainbow Trout", Rarity: "RARE", AwardedAt: daysAgo(12)}, // this month
		{ID: "r-u4-06", UserID: "user4", DisplayName: "DeepDiver", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(8)},  // this month

		// ── user5: 10 rewards; 8 this month → monthly #1 ──
		{ID: "r-u5-01", UserID: "user5", DisplayName: "ReefRider", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(95)},
		{ID: "r-u5-02", UserID: "user5", DisplayName: "ReefRider", Species: "Sunfish", Rarity: "COMMON", AwardedAt: daysAgo(68)},
		{ID: "r-u5-03", UserID: "user5", DisplayName: "ReefRider", Species: "Golden Carp", Rarity: "RARE", AwardedAt: daysAgo(28)}, // this month
		{ID: "r-u5-04", UserID: "user5", DisplayName: "ReefRider", Species: "Blue Tuna", Rarity: "EPIC", AwardedAt: daysAgo(26)},  // this month
		{ID: "r-u5-05", UserID: "user5", DisplayName: "ReefRider", Species: "Phantom Eel", Rarity: "EPIC", AwardedAt: daysAgo(24)}, // this month
		{ID: "r-u5-06", UserID: "user5", DisplayName: "ReefRider", Species: "Rainbow Trout", Rarity: "RARE", AwardedAt: daysAgo(22)}, // this month
		{ID: "r-u5-07", UserID: "user5", DisplayName: "ReefRider", Species: "Glowing Jellyfish", Rarity: "RARE", AwardedAt: daysAgo(18)}, // this month
		{ID: "r-u5-08", UserID: "user5", DisplayName: "ReefRider", Species: "Golden Salmon", Rarity: "RARE", AwardedAt: daysAgo(15)}, // this month
		{ID: "r-u5-09", UserID: "user5", DisplayName: "ReefRider", Species: "Legendary Koi", Rarity: "LEGENDARY", AwardedAt: daysAgo(10)}, // this month
		{ID: "r-u5-10", UserID: "user5", DisplayName: "ReefRider", Species: "Silver Bass", Rarity: "UNCOMMON", AwardedAt: daysAgo(7)}, // this month
	}
}()

// lastUpdate tracks when rewards were last mutated (for leaderboard cache invalidation).
var lastUpdate = time.Now().UTC()

type InMemoryRepository struct {
	mu      sync.RWMutex
	rewards []domain.FishReward
}

func NewInMemory() *InMemoryRepository {
	// Copy seed so concurrent Award() calls don't mutate the package-level slice.
	rewards := make([]domain.FishReward, len(seedRewards))
	copy(rewards, seedRewards)
	return &InMemoryRepository{rewards: rewards}
}

func (repo *InMemoryRepository) Award(ctx context.Context, r *domain.FishReward) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if r.AwardedAt.IsZero() {
		r.AwardedAt = time.Now().UTC()
	}
	repo.rewards = append(repo.rewards, *r)
	lastUpdate = time.Now().UTC()
	return nil
}

func (repo *InMemoryRepository) ListByUser(ctx context.Context, userID string) ([]domain.FishReward, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	var out []domain.FishReward
	for _, r := range repo.rewards {
		if userID == "" || r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

// GetLastUpdate returns the timestamp of the most recent reward mutation.
// The leaderboard service calls this to decide whether its cache is stale.
func (repo *InMemoryRepository) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return lastUpdate, nil
}

