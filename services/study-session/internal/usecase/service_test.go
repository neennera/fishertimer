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

func (r *mockRepository) UpdateSession(ctx context.Context, s *domain.StudySession) error {
	return nil
}

func (r *mockRepository) ListActiveSessions(ctx context.Context) ([]domain.StudySession, error) {
	return []domain.StudySession{{ID: "sess_1", Name: "Deep Work Room", Status: "ACTIVE"}}, nil
}

func (r *mockRepository) AddParticipant(ctx context.Context, p *domain.Participant) error {
	return nil
}

func (r *mockRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	return nil
}

func (r *mockRepository) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	return []domain.Participant{{SessionID: sessionID, UserID: "usr_1"}}, nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	sess, err := svc.CreateSession(ctx, "Test Room", "usr_1", 5)
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}
	if sess.Name != "Test Room" {
		t.Fatalf("expected room name Test Room, got %s", sess.Name)
	}

	ended, err := svc.EndSession(ctx, sess.ID, "usr_1")
	if err != nil {
		t.Fatalf("unexpected error ending session: %v", err)
	}
	if ended.Status != "ENDED" {
		t.Fatalf("expected status ENDED, got %s", ended.Status)
	}

	active, err := svc.ListActiveSession(ctx)
	if err != nil || len(active) == 0 {
		t.Fatalf("expected active sessions, got %v (err: %v)", active, err)
	}

	parts, err := svc.GetParticipants(ctx, "sess_1")
	if err != nil || len(parts) == 0 {
		t.Fatalf("expected participants, got %v (err: %v)", parts, err)
	}
}
