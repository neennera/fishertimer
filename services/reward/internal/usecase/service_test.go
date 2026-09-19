package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

type mockRepository struct{}

func (repo *mockRepository) Award(ctx context.Context, r *domain.FishReward) error {
	return nil
}

func (repo *mockRepository) ListByUser(ctx context.Context, userID string) ([]domain.FishReward, error) {
	return []domain.FishReward{{ID: "fish_1", UserID: userID, Species: "Golden Salmon", Rarity: "RARE"}}, nil
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
