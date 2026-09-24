package repository

import (
	"context"
	"sync"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

type InMemoryRepository struct {
	mu sync.RWMutex
}

func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) ListActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	return []domain.SessionOverview{
		{
			SessionID:        "sess_live_1",
			Name:             "Main Study Room",
			CreatorID:        "usr_101",
			ParticipantCount: 4,
			Status:           "ACTIVE",
			CreatedAt:        time.Now().Add(-30 * time.Minute),
		},
	}, nil
}

func (r *InMemoryRepository) ListParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	return []domain.SessionParticipantOverview{
		{SessionID: sessionID, UserID: "usr_101", JoinedAt: time.Now().Add(-30 * time.Minute)},
		{SessionID: sessionID, UserID: "usr_102", JoinedAt: time.Now().Add(-15 * time.Minute)},
	}, nil
}

func (r *InMemoryRepository) GetSessionDetails(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	return &domain.SessionOverview{
		SessionID:        sessionID,
		Name:             "Main Study Room",
		CreatorID:        "usr_101",
		ParticipantCount: 4,
		Status:           "ACTIVE",
		CreatedAt:        time.Now().Add(-30 * time.Minute),
	}, nil
}
