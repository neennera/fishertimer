package domain

import "errors"

// Reward calculation rules for UC-09 (Award Reward on Cycle Completion).
// Everything here is pure: no database, no clock, and randomness is passed in,
// so the maths can be tested directly against the tables in the project
// description.
const (
	// MinutesPerReward is the length of one full work block that earns a reward (S-1).
	MinutesPerReward = 15
	// BuffPerExtraParticipant is added to the multiplier per participant beyond the first (S-2).
	BuffPerExtraParticipant = 0.25
	// MinParticipants is a solo session: no buff.
	MinParticipants = 1
	// MaxParticipants is the room capacity (UC-01), so the buff tops out at x2.00.
	MaxParticipants = 5
)

// ErrEmptyCatalogue means there is nothing drawable: no items, or every
// adjusted weight is zero.
var ErrEmptyCatalogue = errors.New("reward: catalogue has no drawable items")

// RewardItem is one reward_items catalogue entry.
type RewardItem struct {
	ID         string  `json:"item_id"`
	ItemName   string  `json:"item_name"`
	Category   string  `json:"category"`
	Rarity     string  `json:"rarity"`
	BaseWeight float64 `json:"base_weight"`
	ScoreValue int     `json:"score_value"`
	AssetURL   string  `json:"asset_url"`
}

// RewardCount is S-1: one reward per full 15 minutes of work. Leftover
// minutes are dropped, and a cycle under 15 minutes earns nothing (E-3).
func RewardCount(workMinutes int) int {
	if workMinutes < MinutesPerReward {
		return 0
	}
	return workMinutes / MinutesPerReward
}

// ClampParticipants keeps n inside the range a room can actually hold, so a
// bad caller value can never push the buff past the documented maximum.
func ClampParticipants(n int) int {
	if n < MinParticipants {
		return MinParticipants
	}
	if n > MaxParticipants {
		return MaxParticipants
	}
	return n
}

// CommunityMultiplier is S-2: 1 + 0.25 x (n - 1), with n clamped to 1..5.
func CommunityMultiplier(participants int) float64 {
	n := ClampParticipants(participants)
	return 1 + BuffPerExtraParticipant*float64(n-1)
}

// AdjustedWeight applies the community buff to one item. Common items keep
// their base weight; every other rarity is multiplied.
func AdjustedWeight(item RewardItem, multiplier float64) float64 {
	if item.BaseWeight <= 0 {
		return 0
	}
	if item.Rarity == RarityCommon {
		return item.BaseWeight
	}
	return item.BaseWeight * multiplier
}

// RarityChances returns the chance of each rarity for a single draw, as a
// fraction of 1 (Table 3). Returns nil if nothing is drawable.
func RarityChances(catalogue []RewardItem, participants int) map[string]float64 {
	multiplier := CommunityMultiplier(participants)
	weights := make(map[string]float64)
	total := 0.0
	for _, item := range catalogue {
		w := AdjustedWeight(item, multiplier)
		weights[item.Rarity] += w
		total += w
	}
	if total <= 0 {
		return nil
	}
	for rarity, w := range weights {
		weights[rarity] = w / total
	}
	return weights
}

// DrawRewards rolls the rewards for one completed work cycle. It draws
// RewardCount(workMinutes) items, each independently from the whole catalogue,
// so the same item can come up more than once. An item's chance is its
// adjusted weight over the sum of all adjusted weights.
//
// rnd must return a value in [0, 1), like math/rand's Float64.
func DrawRewards(catalogue []RewardItem, workMinutes, participants int, rnd func() float64) ([]RewardItem, error) {
	count := RewardCount(workMinutes)
	if count == 0 {
		return []RewardItem{}, nil
	}

	multiplier := CommunityMultiplier(participants)
	weights := make([]float64, len(catalogue))
	total := 0.0
	for i, item := range catalogue {
		weights[i] = AdjustedWeight(item, multiplier)
		total += weights[i]
	}
	if total <= 0 {
		return nil, ErrEmptyCatalogue
	}

	drawn := make([]RewardItem, 0, count)
	for i := 0; i < count; i++ {
		drawn = append(drawn, pick(catalogue, weights, rnd()*total))
	}
	return drawn, nil
}

// pick walks the cumulative weights and returns the item whose slice of the
// 0..total line contains target.
func pick(catalogue []RewardItem, weights []float64, target float64) RewardItem {
	last := 0
	for i, w := range weights {
		if w <= 0 {
			continue
		}
		last = i
		if target < w {
			return catalogue[i]
		}
		target -= w
	}
	// Floating point left target a hair past the end: give the last drawable item.
	return catalogue[last]
}
