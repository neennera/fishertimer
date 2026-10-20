package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

func nameOf(t *testing.T, r *domain.CachedRanking, userID string) string {
	t.Helper()
	for _, e := range r.Rankings {
		if e.UserID == userID {
			return e.DisplayName
		}
	}
	t.Fatalf("user %s is not on the ranking", userID)
	return ""
}

func TestRenameUser_ShowsTheNewNameImmediatelyEvenFromACachedRanking(t *testing.T) {
	repo := newMockRepo()
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: seedRewards()}
	svc := usecase.New(repo, rc)
	ctx := context.Background()

	before, err := svc.ViewLeaderboard(ctx, "all-time")
	if err != nil {
		t.Fatal(err)
	}
	if got := nameOf(t, before, "user2"); got != "Beta" {
		t.Fatalf("expected the name on the reward rows, got %q", got)
	}
	// The ranking is now cached and the reward service has not changed.
	if cached, _ := svc.ViewLeaderboard(ctx, "all-time"); !cached.Cached {
		t.Fatal("expected the second view to come from the cache")
	}

	if err := svc.RenameUser(ctx, "user2", "  Beta Prime  "); err != nil {
		t.Fatalf("RenameUser: %v", err)
	}
	after, err := svc.ViewLeaderboard(ctx, "all-time")
	if err != nil {
		t.Fatal(err)
	}
	if got := nameOf(t, after, "user2"); got != "Beta Prime" {
		t.Fatalf("the rename must show without a new reward, got %q", got)
	}
	if got := nameOf(t, after, "user1"); got != "Alpha" {
		t.Fatalf("other users keep their names, got %q", got)
	}
	// Every period is recomputed, not only the one that was cached.
	for _, period := range []string{"weekly", "monthly"} {
		r, err := svc.ViewLeaderboard(ctx, period)
		if err != nil {
			t.Fatal(err)
		}
		if got := nameOf(t, r, "user2"); got != "Beta Prime" {
			t.Fatalf("%s still shows %q", period, got)
		}
	}
}

func TestRenameUser_LastNameWinsAndSurvivesNewRewards(t *testing.T) {
	repo := newMockRepo()
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: seedRewards()}
	svc := usecase.New(repo, rc)
	ctx := context.Background()

	_ = svc.RenameUser(ctx, "user1", "First")
	_ = svc.RenameUser(ctx, "user1", "Second")

	// A new reward arrives under the old name: the remembered name still wins.
	rc.rewards = append(rc.rewards, domain.FishReward{ID: "r9", UserID: "user1", DisplayName: "Alpha", ScoreValue: 10, AwardedAt: time.Now().UTC()})
	rc.lastUpdate = time.Now().UTC().Add(time.Minute)

	r, err := svc.ViewLeaderboard(ctx, "all-time")
	if err != nil {
		t.Fatal(err)
	}
	if got := nameOf(t, r, "user1"); got != "Second" {
		t.Fatalf("expected the latest rename, got %q", got)
	}
}

func TestRenameUser_RejectsUnusableInput(t *testing.T) {
	svc := usecase.New(newMockRepo(), &mockRewardClient{lastUpdate: time.Now().UTC()})
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"no user":   func() error { return svc.RenameUser(ctx, "  ", "Name") },
		"no name":   func() error { return svc.RenameUser(ctx, "user1", "   ") },
		"too long":  func() error { return svc.RenameUser(ctx, "user1", strings.Repeat("x", 101)) },
		"all empty": func() error { return svc.RenameUser(ctx, "", "") },
	} {
		if err := call(); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: expected ErrInvalid, got %v", name, err)
		}
	}
}

// A rename that lands while a ranking is being computed must not leave the old
// name cached: the next view has to show the new one.
//
// The rename is made to arrive right after the remembered names were read, the
// moment the ranking is already committed to the old name.
type renameAfterNamesRead struct {
	*mockRepository
	afterRead func()
}

func (r *renameAfterNamesRead) GetDisplayNames(ctx context.Context) (map[string]string, error) {
	names, err := r.mockRepository.GetDisplayNames(ctx)
	if r.afterRead != nil {
		hook := r.afterRead
		r.afterRead = nil
		hook()
	}
	return names, err
}

func TestViewLeaderboard_DoesNotCacheARankingComputedAcrossARename(t *testing.T) {
	repo := &renameAfterNamesRead{mockRepository: newMockRepo()}
	rc := &mockRewardClient{lastUpdate: time.Now().UTC(), rewards: seedRewards()}
	svc := usecase.New(repo, rc)
	ctx := context.Background()

	repo.afterRead = func() {
		// What the consumer does for profile.updated, mid-compute.
		_ = svc.RenameUser(ctx, "user2", "Beta Prime")
	}
	first, err := svc.ViewLeaderboard(ctx, "all-time")
	if err != nil {
		t.Fatal(err)
	}
	if got := nameOf(t, first, "user2"); got != "Beta" {
		t.Fatalf("the in-flight view was computed with the old name, got %q", got)
	}

	next, err := svc.ViewLeaderboard(ctx, "all-time")
	if err != nil {
		t.Fatal(err)
	}
	if got := nameOf(t, next, "user2"); got != "Beta Prime" {
		t.Fatalf("a stale ranking was cached across the rename: got %q", got)
	}
}
