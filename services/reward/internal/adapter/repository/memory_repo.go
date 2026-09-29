package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

func seedCatch(id, userID, displayName, itemName, rarity string, awardedAt time.Time) domain.UnlockedReward {
	var score int
	var weight float64
	switch rarity {
	case domain.RarityLegendary:
		score = 250
		weight = 1
	case domain.RarityEpic:
		score = 100
		weight = 4
	case domain.RarityRare:
		score = 50
		weight = 10
	case domain.RarityUncommon:
		score = 25
		weight = 25
	default:
		score = 10
		weight = 50
	}

	sprite := "Bass.png"
	switch itemName {
	case "Legendary Koi", "Koi":
		sprite = "Koi.png"
	case "Rainbow Trout":
		sprite = "Rainbow Trout.png"
	case "Golden Salmon", "Golden Carp", "Goldfish":
		sprite = "Goldfish.png"
	case "Blue Tuna", "Phantom Eel", "Ghostfish":
		sprite = "Ghostfish.png"
	case "Glowing Jellyfish", "Angelfish":
		sprite = "Angelfish.png"
	case "Catfish":
		sprite = "Catfish.png"
	case "Clownfish":
		sprite = "Clownfish.png"
	case "Octopus":
		sprite = "Octopus.png"
	case "Seahorse":
		sprite = "Seahorse.png"
	case "Globefish":
		sprite = "Globefish.png"
	}

	return domain.UnlockedReward{
		UserRewardID: id,
		ID:           id,
		ItemID:       "item-" + id,
		UserID:       userID,
		DisplayName:  displayName,
		ItemName:     itemName,
		Species:      itemName,
		Category:     domain.CategoryFish,
		Rarity:       rarity,
		ScoreValue:   score,
		BaseWeight:   weight,
		AssetURL:     "/sprites/fish/" + sprite,
		AwardedAt:    awardedAt,
	}
}

// Seed timestamps — all relative to "now" so period filters are always live.
//
//	user1: 12 rewards over 3 months  → all-time #1
//	user2:  8 rewards (2 this week)  → weekly #2
//	user3:  5 rewards (5 this week)  → weekly #1, monthly top-5
//	user4:  6 rewards (3 this month) → monthly contender
//	user5: 10 rewards (8 this month) → monthly #1
func generateSeedRewards() []domain.UnlockedReward {
	now := time.Now().UTC()

	// helper: returns a time N days ago
	daysAgo := func(n int) time.Time {
		return now.AddDate(0, 0, -n)
	}

	return []domain.UnlockedReward{
		// ── user1: 12 rewards spread over ~90 days (all-time king) ──
		seedCatch("r-u1-01", "user1", "TideAngler", "Legendary Koi", domain.RarityLegendary, daysAgo(90)),
		seedCatch("r-u1-02", "user1", "TideAngler", "Silver Bass", domain.RarityUncommon, daysAgo(80)),
		seedCatch("r-u1-03", "user1", "TideAngler", "Rainbow Trout", domain.RarityRare, daysAgo(70)),
		seedCatch("r-u1-04", "user1", "TideAngler", "Golden Carp", domain.RarityRare, daysAgo(60)),
		seedCatch("r-u1-05", "user1", "TideAngler", "Phantom Eel", domain.RarityEpic, daysAgo(50)),
		seedCatch("r-u1-06", "user1", "TideAngler", "Silver Bass", domain.RarityUncommon, daysAgo(42)),
		seedCatch("r-u1-07", "user1", "TideAngler", "Sunfish", domain.RarityCommon, daysAgo(35)),
		seedCatch("r-u1-08", "user1", "TideAngler", "River Perch", domain.RarityCommon, daysAgo(28)),
		seedCatch("r-u1-09", "user1", "TideAngler", "Blue Tuna", domain.RarityEpic, daysAgo(21)),
		seedCatch("r-u1-10", "user1", "TideAngler", "Glowing Jellyfish", domain.RarityRare, daysAgo(14)),
		seedCatch("r-u1-11", "user1", "TideAngler", "Sunfish", domain.RarityCommon, daysAgo(10)),
		seedCatch("r-u1-12", "user1", "TideAngler", "River Perch", domain.RarityCommon, daysAgo(9)),

		// ── user2: 8 rewards; 2 this week ──
		seedCatch("r-u2-01", "user2", "CastMaster", "Golden Salmon", domain.RarityRare, daysAgo(75)),
		seedCatch("r-u2-02", "user2", "CastMaster", "Phantom Eel", domain.RarityEpic, daysAgo(55)),
		seedCatch("r-u2-03", "user2", "CastMaster", "River Perch", domain.RarityCommon, daysAgo(40)),
		seedCatch("r-u2-04", "user2", "CastMaster", "Sunfish", domain.RarityCommon, daysAgo(30)),
		seedCatch("r-u2-05", "user2", "CastMaster", "Rainbow Trout", domain.RarityRare, daysAgo(22)),
		seedCatch("r-u2-06", "user2", "CastMaster", "Silver Bass", domain.RarityUncommon, daysAgo(15)),
		seedCatch("r-u2-07", "user2", "CastMaster", "Blue Tuna", domain.RarityEpic, daysAgo(3)),  // this week
		seedCatch("r-u2-08", "user2", "CastMaster", "River Perch", domain.RarityCommon, daysAgo(1)), // this week

		// ── user3: 5 rewards ALL this week → weekly #1 ──
		seedCatch("r-u3-01", "user3", "LureQueen", "Glowing Jellyfish", domain.RarityRare, daysAgo(6)),
		seedCatch("r-u3-02", "user3", "LureQueen", "Rainbow Trout", domain.RarityRare, daysAgo(5)),
		seedCatch("r-u3-03", "user3", "LureQueen", "Golden Carp", domain.RarityRare, daysAgo(4)),
		seedCatch("r-u3-04", "user3", "LureQueen", "Blue Tuna", domain.RarityEpic, daysAgo(2)),
		seedCatch("r-u3-05", "user3", "LureQueen", "Phantom Eel", domain.RarityEpic, daysAgo(1)),

		// ── user4: 6 rewards; 3 this month ──
		seedCatch("r-u4-01", "user4", "DeepDiver", "Sunfish", domain.RarityCommon, daysAgo(85)),
		seedCatch("r-u4-02", "user4", "DeepDiver", "Silver Bass", domain.RarityUncommon, daysAgo(65)),
		seedCatch("r-u4-03", "user4", "DeepDiver", "River Perch", domain.RarityCommon, daysAgo(45)),
		seedCatch("r-u4-04", "user4", "DeepDiver", "Golden Salmon", domain.RarityRare, daysAgo(20)), // this month
		seedCatch("r-u4-05", "user4", "DeepDiver", "Rainbow Trout", domain.RarityRare, daysAgo(12)), // this month
		seedCatch("r-u4-06", "user4", "DeepDiver", "Sunfish", domain.RarityCommon, daysAgo(8)),  // this month

		// ── user5: 10 rewards; 8 this month → monthly #1 ──
		seedCatch("r-u5-01", "user5", "ReefRider", "Silver Bass", domain.RarityUncommon, daysAgo(95)),
		seedCatch("r-u5-02", "user5", "ReefRider", "Sunfish", domain.RarityCommon, daysAgo(68)),
		seedCatch("r-u5-03", "user5", "ReefRider", "Golden Carp", domain.RarityRare, daysAgo(28)), // this month
		seedCatch("r-u5-04", "user5", "ReefRider", "Blue Tuna", domain.RarityEpic, daysAgo(26)),  // this month
		seedCatch("r-u5-05", "user5", "ReefRider", "Phantom Eel", domain.RarityEpic, daysAgo(24)), // this month
		seedCatch("r-u5-06", "user5", "ReefRider", "Rainbow Trout", domain.RarityRare, daysAgo(22)), // this month
		seedCatch("r-u5-07", "user5", "ReefRider", "Glowing Jellyfish", domain.RarityRare, daysAgo(18)), // this month
		seedCatch("r-u5-08", "user5", "ReefRider", "Golden Salmon", domain.RarityRare, daysAgo(15)), // this month
		seedCatch("r-u5-09", "user5", "ReefRider", "Legendary Koi", domain.RarityLegendary, daysAgo(10)), // this month
		seedCatch("r-u5-10", "user5", "ReefRider", "Silver Bass", domain.RarityUncommon, daysAgo(7)), // this month
	}
}

// lastUpdate tracks when rewards were last mutated (for leaderboard cache invalidation).
var lastUpdate = time.Now().UTC()

type InMemoryRepository struct {
	mu      sync.RWMutex
	rewards []domain.UnlockedReward
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{rewards: generateSeedRewards()}
}

func (repo *InMemoryRepository) Award(ctx context.Context, r *domain.UnlockedReward) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if r.AwardedAt.IsZero() {
		r.AwardedAt = time.Now().UTC()
	}
	if r.UserRewardID == "" {
		r.UserRewardID = fmt.Sprintf("r-%d", time.Now().UnixNano())
	}
	if r.ID == "" {
		r.ID = r.UserRewardID
	}
	repo.rewards = append(repo.rewards, *r)
	lastUpdate = time.Now().UTC()
	return nil
}

func (repo *InMemoryRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	var out []domain.UnlockedReward
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
