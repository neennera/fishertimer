package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

type mockRepository struct{}

func (r *mockRepository) ListActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	return []domain.SessionOverview{{SessionID: "s1", Name: "Room 1", Status: "ACTIVE", CreatedAt: time.Now()}}, nil
}

func (r *mockRepository) ListParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	return []domain.SessionParticipantOverview{{SessionID: sessionID, UserID: "u1", JoinedAt: time.Now()}}, nil
}

func (r *mockRepository) GetSessionDetails(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	return &domain.SessionOverview{SessionID: sessionID, Name: "Room 1", Status: "ACTIVE"}, nil
}

type mockSessionClient struct {
	kicked bool
	closed bool
}

func (m *mockSessionClient) KickParticipant(ctx context.Context, sessionID, userID string) error {
	m.kicked = true
	return nil
}

func (m *mockSessionClient) CloseSession(ctx context.Context, sessionID string) error {
	m.closed = true
	return nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	client := &mockSessionClient{}
	svc := usecase.New(repo, client)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	sessions, err := svc.ViewActiveSessions(ctx)
	if err != nil || len(sessions) == 0 {
		t.Fatalf("unexpected error viewing sessions: %v", err)
	}

	parts, err := svc.ViewParticipants(ctx, "s1")
	if err != nil || len(parts) == 0 {
		t.Fatalf("unexpected error viewing participants: %v", err)
	}

	err = svc.KickUser(ctx, "s1", "u1")
	if err != nil || !client.kicked {
		t.Fatalf("unexpected error kicking user: %v", err)
	}

	err = svc.EndSession(ctx, "s1")
	if err != nil || !client.closed {
		t.Fatalf("unexpected error ending session: %v", err)
	}
}
