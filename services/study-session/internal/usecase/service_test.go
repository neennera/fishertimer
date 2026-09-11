package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) GetSession(ctx context.Context, id string) (*domain.StudySession, error) {
	return &domain.StudySession{ID: id, Name: "Deep Work Room", ParticipantLimit: 10, Status: "ACTIVE"}, nil
}

func (r *mockRepository) CreateSession(ctx context.Context, s *domain.StudySession) error {
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
