package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	return &domain.TimerState{SessionID: sessionID, UserID: userID, Status: "STOPPED"}, nil
}

func (r *mockRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
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
