package repository

import "testing"

func TestAwardedRewardID_DeterministicPerDraw(t *testing.T) {
	if awardedRewardID("cycle-1", 0) != awardedRewardID("cycle-1", 0) {
		t.Fatal("same cycle and draw must give the same _id, so a retry collides")
	}

	seen := map[string]int{}
	for n := 0; n < 8; n++ {
		id := awardedRewardID("cycle-1", n).Hex()
		if prev, ok := seen[id]; ok {
			t.Fatalf("draws %d and %d share _id %s", prev, n, id)
		}
		seen[id] = n
	}

	if awardedRewardID("cycle-1", 0) == awardedRewardID("cycle-2", 0) {
		t.Fatal("different cycles must not share an _id")
	}
}
