package domain_test

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

// docCatalogue is Table 2 of the project description: items per rarity and
// base weight per item. It matches database/schemas/002_seed_reward_items.js.
func docCatalogue() []domain.RewardItem {
	tiers := []struct {
		rarity string
		items  int
		weight float64
		score  int
	}{
		{domain.RarityCommon, 5, 30, 10},
		{domain.RarityUncommon, 2, 25, 25},
		{domain.RarityRare, 3, 12, 50},
		{domain.RarityEpic, 3, 6, 100},
		{domain.RarityLegendary, 2, 3, 250},
	}
	var catalogue []domain.RewardItem
	for _, tier := range tiers {
		for i := 1; i <= tier.items; i++ {
			catalogue = append(catalogue, domain.RewardItem{
				ID:         fmt.Sprintf("%s-%d", tier.rarity, i),
				ItemName:   fmt.Sprintf("%s %d", tier.rarity, i),
				Category:   domain.CategoryFish,
				Rarity:     tier.rarity,
				BaseWeight: tier.weight,
				ScoreValue: tier.score,
			})
		}
	}
	return catalogue
}

// Table 1: work cycle duration -> reward count.
func TestRewardCount_Table1(t *testing.T) {
	cases := []struct {
		minutes, want int
	}{
		{-5, 0}, {0, 0}, {1, 0}, {14, 0},
		{15, 1}, {29, 1},
		{30, 2}, {44, 2},
		{45, 3}, {59, 3},
		{60, 4}, {74, 4},
		{75, 5},
	}
	for _, c := range cases {
		if got := domain.RewardCount(c.minutes); got != c.want {
			t.Errorf("RewardCount(%d) = %d, want %d", c.minutes, got, c.want)
		}
	}
}

// S-2: multiplier = 1 + 0.25 x (n - 1), clamped to the room capacity.
func TestCommunityMultiplier(t *testing.T) {
	cases := []struct {
		participants int
		want         float64
	}{
		{1, 1.00}, {2, 1.25}, {3, 1.50}, {4, 1.75}, {5, 2.00},
		{0, 1.00}, {-3, 1.00}, // below a solo session: no buff
		{6, 2.00}, {100, 2.00}, // above capacity: capped
	}
	for _, c := range cases {
		if got := domain.CommunityMultiplier(c.participants); got != c.want {
			t.Errorf("CommunityMultiplier(%d) = %v, want %v", c.participants, got, c.want)
		}
	}
}

// Table 3: chance of each rarity for one draw, by participant count.
func TestRarityChances_Table3(t *testing.T) {
	rarities := []string{
		domain.RarityCommon, domain.RarityUncommon, domain.RarityRare,
		domain.RarityEpic, domain.RarityLegendary,
	}
	// Percentages exactly as printed in the document (one decimal place).
	table := map[int][]float64{
		1: {57.7, 19.2, 13.8, 6.9, 2.3},
		2: {52.2, 21.7, 15.7, 7.8, 2.6},
		3: {47.6, 23.8, 17.1, 8.6, 2.9},
		4: {43.8, 25.5, 18.4, 9.2, 3.1},
		5: {40.5, 27.0, 19.5, 9.7, 3.2},
	}
	catalogue := docCatalogue()
	for n, want := range table {
		chances := domain.RarityChances(catalogue, n)
		sum := 0.0
		for i, rarity := range rarities {
			got := chances[rarity] * 100
			sum += got
			// The document rounds to 0.1%, so allow half of that.
			if math.Abs(got-want[i]) > 0.05+1e-9 {
				t.Errorf("n=%d %s = %.3f%%, want %.1f%%", n, rarity, got, want[i])
			}
		}
		if math.Abs(sum-100) > 1e-9 {
			t.Errorf("n=%d chances sum to %.6f%%, want 100%%", n, sum)
		}
	}
}

// The worked example in UC-09: 60 minutes, 5 participants.
func TestWorkedExample_60MinutesFiveParticipants(t *testing.T) {
	catalogue := docCatalogue()

	if got := domain.RewardCount(60); got != 4 {
		t.Fatalf("reward count = %d, want 4", got)
	}
	multiplier := domain.CommunityMultiplier(5)
	if multiplier != 2.00 {
		t.Fatalf("multiplier = %v, want 2.00", multiplier)
	}

	baseTotal, buffedTotal := 0.0, 0.0
	for _, item := range catalogue {
		baseTotal += domain.AdjustedWeight(item, 1)
		buffedTotal += domain.AdjustedWeight(item, multiplier)
	}
	if baseTotal != 260 || buffedTotal != 370 {
		t.Fatalf("weight totals = %v and %v, want 260 and 370", baseTotal, buffedTotal)
	}

	rare := domain.RewardItem{Rarity: domain.RarityRare, BaseWeight: 12}
	if got := domain.AdjustedWeight(rare, multiplier); got != 24 {
		t.Errorf("rare item adjusted weight = %v, want 24", got)
	}
	common := domain.RewardItem{Rarity: domain.RarityCommon, BaseWeight: 30}
	if got := domain.AdjustedWeight(common, multiplier); got != 30 {
		t.Errorf("common item adjusted weight = %v, want 30 (never buffed)", got)
	}

	drawn, err := domain.DrawRewards(catalogue, 60, 5, rand.New(rand.NewSource(1)).Float64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drawn) != 4 {
		t.Fatalf("drew %d rewards, want 4", len(drawn))
	}
}

// E-3: under 15 minutes gives nothing, and never touches the random source.
func TestDrawRewards_ShortCycleGivesNothing(t *testing.T) {
	rnd := func() float64 {
		t.Fatal("random source must not be used when no reward is due")
		return 0
	}
	drawn, err := domain.DrawRewards(docCatalogue(), 14, 5, rnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drawn) != 0 {
		t.Fatalf("drew %d rewards, want 0", len(drawn))
	}
}

// Fixed random values land on known items: the draw walks cumulative weights.
func TestDrawRewards_PicksByCumulativeWeight(t *testing.T) {
	catalogue := []domain.RewardItem{
		{ID: "common", Rarity: domain.RarityCommon, BaseWeight: 30},
		{ID: "rare", Rarity: domain.RarityRare, BaseWeight: 10},
	}
	// Solo: weights 30 | 10, total 40.   common = [0, 0.75), rare = [0.75, 1)
	// n=5:  weights 30 | 20, total 50.   common = [0, 0.60), rare = [0.60, 1)
	cases := []struct {
		participants int
		roll         float64
		want         string
	}{
		{1, 0.00, "common"}, {1, 0.74, "common"}, {1, 0.75, "rare"}, {1, 0.999999, "rare"},
		{5, 0.59, "common"}, {5, 0.60, "rare"}, {5, 0.74, "rare"},
	}
	for _, c := range cases {
		drawn, err := domain.DrawRewards(catalogue, 15, c.participants, func() float64 { return c.roll })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(drawn) != 1 || drawn[0].ID != c.want {
			t.Errorf("n=%d roll=%v drew %+v, want %s", c.participants, c.roll, drawn, c.want)
		}
	}
}

// Draws are independent, so one cycle can return the same item twice.
func TestDrawRewards_SameItemCanRepeat(t *testing.T) {
	drawn, err := domain.DrawRewards(docCatalogue(), 60, 1, func() float64 { return 0 })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drawn) != 4 {
		t.Fatalf("drew %d rewards, want 4", len(drawn))
	}
	for _, item := range drawn {
		if item.ID != drawn[0].ID {
			t.Fatalf("expected the same item four times, got %+v", drawn)
		}
	}
}

func TestDrawRewards_ZeroWeightItemIsNeverDrawn(t *testing.T) {
	catalogue := []domain.RewardItem{
		{ID: "retired", Rarity: domain.RarityCommon, BaseWeight: 0},
		{ID: "live", Rarity: domain.RarityCommon, BaseWeight: 5},
		{ID: "retired-too", Rarity: domain.RarityLegendary, BaseWeight: 0},
	}
	rng := rand.New(rand.NewSource(7))
	drawn, err := domain.DrawRewards(catalogue, 15*200, 5, rng.Float64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, item := range drawn {
		if item.ID != "live" {
			t.Fatalf("drew %q, which has zero weight", item.ID)
		}
	}
}

func TestDrawRewards_EmptyCatalogue(t *testing.T) {
	for name, catalogue := range map[string][]domain.RewardItem{
		"no items":         nil,
		"all zero weights": {{ID: "a", Rarity: domain.RarityRare, BaseWeight: 0}},
	} {
		_, err := domain.DrawRewards(catalogue, 30, 1, func() float64 { return 0.5 })
		if !errors.Is(err, domain.ErrEmptyCatalogue) {
			t.Errorf("%s: err = %v, want ErrEmptyCatalogue", name, err)
		}
	}
}

// Statistical check: over many draws the observed rarity share matches
// Table 3. Seeded, so it is repeatable.
func TestDrawRewards_DistributionMatchesTable3(t *testing.T) {
	const draws = 200_000
	catalogue := docCatalogue()
	for _, n := range []int{1, 3, 5} {
		rng := rand.New(rand.NewSource(int64(42 + n)))
		drawn, err := domain.DrawRewards(catalogue, draws*domain.MinutesPerReward, n, rng.Float64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(drawn) != draws {
			t.Fatalf("drew %d rewards, want %d", len(drawn), draws)
		}
		seen := map[string]int{}
		for _, item := range drawn {
			seen[item.Rarity]++
		}
		for rarity, want := range domain.RarityChances(catalogue, n) {
			got := float64(seen[rarity]) / draws
			if math.Abs(got-want) > 0.005 { // within half a percentage point
				t.Errorf("n=%d %s observed %.2f%%, expected %.2f%%", n, rarity, got*100, want*100)
			}
		}
	}
}
