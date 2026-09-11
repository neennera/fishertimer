package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return &domain.Profile{UserID: userID, DisplayName: "Fisher"}, nil
}

func (r *mockRepository) UpdateBanStatus(ctx context.Context, userID string, banned bool) error {
	return nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()
	_ = ctx
}
