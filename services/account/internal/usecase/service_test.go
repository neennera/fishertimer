package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	return nil
}

func (r *mockRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	return &domain.UserAccount{UserID: "usr_1", Email: email, DisplayName: "Fisher"}, nil
}

func (r *mockRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return &domain.Profile{UserID: userID, DisplayName: "Fisher"}, nil
}

func (r *mockRepository) UpdateProfile(ctx context.Context, p *domain.Profile) error {
	return nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	u, err := svc.SignIn(ctx, "test@example.com", "Fisher")
	if err != nil || u.Email != "test@example.com" {
		t.Fatalf("unexpected signIn error: %v", err)
	}

	p, err := svc.ViewProfile(ctx, "usr_1")
	if err != nil || p.DisplayName != "Fisher" {
		t.Fatalf("unexpected viewProfile error: %v", err)
	}

	stats, err := svc.ViewStatistics(ctx, "usr_1")
	if err != nil || stats.TotalSessions <= 0 {
		t.Fatalf("unexpected viewStats error: %v", err)
	}
}
