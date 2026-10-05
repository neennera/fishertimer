package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

// mockRepository keeps rows in memory and, like the real adapters, rejects a
// second award for a cycle_id.
type mockRepository struct {
	catalogue []domain.RewardItem
	rows      []domain.UnlockedReward
	// beforeAward runs at the start of AwardMany, to simulate a concurrent
	// request storing the same cycle first.
	beforeAward func(repo *mockRepository)
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		catalogue: []domain.RewardItem{
			{ID: "fish-common", ItemName: "Bass", Category: domain.CategoryFish, Rarity: domain.RarityCommon, BaseWeight: 30, ScoreValue: 10, AssetURL: "/sprites/fish/Bass.png"},
			{ID: "fish-rare", ItemName: "Pufferfish", Category: domain.CategoryFish, Rarity: domain.RarityRare, BaseWeight: 12, ScoreValue: 50, AssetURL: "/sprites/fish/Pufferfish.png"},
		},
		rows: []domain.UnlockedReward{
			{UserRewardID: "seed-1", ItemID: "fish-common", UserID: "usr_123", CycleID: "c-seed", ItemName: "Bass", Category: domain.CategoryFish},
		},
	}
}

func (repo *mockRepository) ListItems(ctx context.Context) ([]domain.RewardItem, error) {
	return repo.catalogue, nil
}

func (repo *mockRepository) ListByCycle(ctx context.Context, cycleID string) ([]domain.UnlockedReward, error) {
	var out []domain.UnlockedReward
	for _, r := range repo.rows {
		if r.CycleID == cycleID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (repo *mockRepository) AwardMany(ctx context.Context, rewards []domain.UnlockedReward) error {
	if repo.beforeAward != nil {
		repo.beforeAward(repo)
	}
	for _, r := range rewards {
		for _, row := range repo.rows {
			if row.CycleID == r.CycleID {
				return domain.ErrAlreadyAwarded
			}
		}
	}
	repo.rows = append(repo.rows, rewards...)
	return nil
}

func (repo *mockRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	var out []domain.UnlockedReward
	for _, r := range repo.rows {
		if userID == "" || r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (repo *mockRepository) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

// awardedRows counts the rows stored for a cycle by AwardReward.
func (repo *mockRepository) awardedRows(cycleID string) int {
	rows, _ := repo.ListByCycle(context.Background(), cycleID)
	return len(rows)
}

// fixedRand always lands on the first catalogue item.
func fixedRand() float64 { return 0 }

func award(cycleID string, workMinutes int) usecase.AwardInput {
	return usecase.AwardInput{
		UserID:           "usr_123",
		SessionID:        "s-1",
		CycleID:          cycleID,
		WorkMinutes:      workMinutes,
		ParticipantCount: 1,
		DisplayName:      "Angler",
	}
}

func TestAwardReward_SixtyMinutesStoresFourRows(t *testing.T) {
	repo := newMockRepository()
	svc := usecase.NewWithRand(repo, fixedRand)

	result, err := svc.AwardReward(context.Background(), award("c-1", 60))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AlreadyAwarded {
		t.Fatal("first award should not be already_awarded")
	}
	if len(result.Rewards) != 4 || repo.awardedRows("c-1") != 4 {
		t.Fatalf("expected 4 rewards and 4 rows, got %d and %d", len(result.Rewards), repo.awardedRows("c-1"))
	}
	for i, r := range result.Rewards {
		if r.ItemID != "fish-common" || r.ItemName != "Bass" || r.Species != "Bass" ||
			r.Rarity != domain.RarityCommon || r.ScoreValue != 10 || r.AssetURL != "/sprites/fish/Bass.png" {
			t.Fatalf("reward %d not filled from the catalogue: %+v", i, r)
		}
		if r.UserID != "usr_123" || r.CycleID != "c-1" || r.DisplayName != "Angler" {
			t.Fatalf("reward %d has the wrong owner fields: %+v", i, r)
		}
	}
}

func TestAwardReward_SameCycleTwiceIsAlreadyAwarded(t *testing.T) {
	repo := newMockRepository()
	svc := usecase.NewWithRand(repo, fixedRand)
	ctx := context.Background()

	if _, err := svc.AwardReward(ctx, award("c-1", 60)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := svc.AwardReward(ctx, award("c-1", 60))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AlreadyAwarded {
		t.Fatal("expected already_awarded on the retry")
	}
	if len(result.Rewards) != 4 || repo.awardedRows("c-1") != 4 {
		t.Fatalf("expected the same 4 rewards and still 4 rows, got %d and %d", len(result.Rewards), repo.awardedRows("c-1"))
	}
}

func TestAwardReward_ConcurrentRetryReturnsWinnersRewards(t *testing.T) {
	repo := newMockRepository()
	svc := usecase.NewWithRand(repo, fixedRand)
	ctx := context.Background()

	// Another request stores c-1 between this request's check and insert.
	repo.beforeAward = func(r *mockRepository) {
		r.beforeAward = nil
		if _, err := usecase.NewWithRand(r, fixedRand).AwardReward(ctx, award("c-1", 30)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	result, err := svc.AwardReward(ctx, award("c-1", 60))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AlreadyAwarded {
		t.Fatal("expected already_awarded when the insert hits a duplicate")
	}
	if len(result.Rewards) != 2 || repo.awardedRows("c-1") != 2 {
		t.Fatalf("expected the winner's 2 rewards and 2 rows, got %d and %d", len(result.Rewards), repo.awardedRows("c-1"))
	}
}

func TestAwardReward_UnderFifteenMinutesStoresNothing(t *testing.T) {
	repo := newMockRepository()
	svc := usecase.NewWithRand(repo, fixedRand)

	result, err := svc.AwardReward(context.Background(), award("c-1", 14))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AlreadyAwarded || len(result.Rewards) != 0 || repo.awardedRows("c-1") != 0 {
		t.Fatalf("expected no rewards and no rows, got %+v and %d rows", result, repo.awardedRows("c-1"))
	}
	if result.Rewards == nil {
		t.Fatal("expected an empty list, not nil, so the JSON is []")
	}
}

func TestAwardReward_InvalidInput(t *testing.T) {
	svc := usecase.NewWithRand(newMockRepository(), fixedRand)

	cases := map[string]usecase.AwardInput{
		"empty cycle_id":        award("", 60),
		"empty user_id":         {CycleID: "c-1", WorkMinutes: 60},
		"negative work_minutes": award("c-1", -1),
	}
	for name, in := range cases {
		if _, err := svc.AwardReward(context.Background(), in); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("%s: expected ErrInvalid, got %v", name, err)
		}
	}
}

func TestUsecase_ListsRewards(t *testing.T) {
	svc := usecase.New(newMockRepository())
	ctx := context.Background()

	items, err := svc.GetUserInventory(ctx, "usr_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected items, got none")
	}

	all, err := svc.ListAllRewards(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected all rewards, got none")
	}

	if _, err := svc.GetLastUpdate(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
