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

type mockRewardClient struct {
	called bool
	userID string
	reason string
}

func (m *mockRewardClient) AwardReward(ctx context.Context, userID, reason string) error {
	m.called = true
	m.userID = userID
	m.reason = reason
	return nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	rewardClient := &mockRewardClient{}
	svc := usecase.New(repo, rewardClient)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	sess, err := svc.EndSession(ctx, "sess_1", "usr_999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.Status != "ENDED" {
		t.Fatalf("expected status ENDED, got %s", sess.Status)
	}
	if !rewardClient.called {
		t.Fatal("expected rewardClient.AwardReward to be called upon EndSession")
	}
	if rewardClient.userID != "usr_999" {
		t.Fatalf("expected userID usr_999, got %s", rewardClient.userID)
	}
}
