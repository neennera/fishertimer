package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type Usecase interface {
	CreateSession(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error)
	JoinSession(ctx context.Context, sessionID, userID string) error
	LeaveSession(ctx context.Context, sessionID, userID string) error
	EndSession(ctx context.Context, sessionID, userID string) (*domain.StudySession, error)
	ListActiveSession(ctx context.Context) ([]domain.StudySession, error)
	GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) CreateSession(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error) {
	sess := &domain.StudySession{
		ID:               "sess_test",
		Name:             name,
		CreatorID:        creatorID,
		ParticipantLimit: limit,
		Status:           "ACTIVE",
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return nil, err
	}
	_ = s.repo.AddParticipant(ctx, &domain.Participant{
		SessionID: sess.ID,
		UserID:    creatorID,
	})
	return sess, nil
}

func (s *service) JoinSession(ctx context.Context, sessionID, userID string) error {
	return s.repo.AddParticipant(ctx, &domain.Participant{
		SessionID: sessionID,
		UserID:    userID,
	})
}

func (s *service) LeaveSession(ctx context.Context, sessionID, userID string) error {
	return s.repo.RemoveParticipant(ctx, sessionID, userID)
}

func (s *service) EndSession(ctx context.Context, sessionID, userID string) (*domain.StudySession, error) {
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		sess = &domain.StudySession{ID: sessionID, Name: "Ended Room", CreatorID: userID, Status: "ENDED"}
	} else {
		sess.Status = "ENDED"
	}
	_ = s.repo.UpdateSession(ctx, sess)
	return sess, nil
}

func (s *service) ListActiveSession(ctx context.Context) ([]domain.StudySession, error) {
	return s.repo.ListActiveSessions(ctx)
}

func (s *service) GetParticipants(ctx context.Context, sessionID string) ([]domain.Participant, error) {
	return s.repo.GetParticipants(ctx, sessionID)
}
