package usecase_test

import (
	"context"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
	"testing"
)

type mockRepository struct{}

func (repo *mockRepository) Award(ctx context.Context, r *domain.UnlockedReward) error {
	return nil
}

func (repo *mockRepository) ListByUser(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return []domain.UnlockedReward{{ItemID: "fish_1", UserID: userID, ItemName: "Golden Salmon", Category: domain.CategoryFish}}, nil
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
}
