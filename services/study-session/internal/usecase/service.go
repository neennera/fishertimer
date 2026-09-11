package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type Usecase interface {
	CreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error)
	GetRoom(ctx context.Context, id string) (*domain.StudySession, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) CreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error) {
	sess := &domain.StudySession{ID: "sess_test", Name: name, CreatorID: creatorID, ParticipantLimit: limit, Status: "ACTIVE"}
	return sess, s.repo.CreateSession(ctx, sess)
}

func (s *service) GetRoom(ctx context.Context, id string) (*domain.StudySession, error) {
	return s.repo.GetSession(ctx, id)
}
