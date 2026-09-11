package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/auth/internal/domain"
	"github.com/neennera/fishertimer/services/auth/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) FindByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	return nil, domain.ErrNotFound
}

func (r *mockRepository) Create(ctx context.Context, user *domain.UserAccount) error {
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
