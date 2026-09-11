package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

type mockRepository struct{}

func (repo *mockRepository) CreateReport(ctx context.Context, r *domain.ModerationReport) error {
	return nil
}

func (repo *mockRepository) ListPending(ctx context.Context) ([]domain.ModerationReport, error) {
	return []domain.ModerationReport{{ID: "rep_1", Reason: "AFK Spam", Status: "PENDING"}}, nil
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
