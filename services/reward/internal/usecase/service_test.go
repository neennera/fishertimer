package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

type mockRepository struct{}

func (repo *mockRepository) Award(ctx context.Context, r *domain.UnlockedReward) error {
	return nil
}

func (repo *mockRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return []domain.UnlockedReward{{ItemID: "fish_1", UserID: userID, ItemName: "Golden Salmon", Category: domain.CategoryFish}}, nil
}

func (repo *mockRepository) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	reward, err := svc.AwardReward(ctx, "usr_123", "session_ended")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward.UserID != "usr_123" {
		t.Fatalf("expected user_id usr_123, got %s", reward.UserID)
	}

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

	_, err = svc.GetLastUpdate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsecase_InvalidUser(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)

	_, err := svc.AwardReward(context.Background(), "", "session_ended")
	if err != domain.ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
