package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

type Usecase interface {
	CreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error)
	GetRoom(ctx context.Context, id string) (*domain.StudySession, error)
	EndSession(ctx context.Context, sessionID, userID string) (*domain.StudySession, error)
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient
}

func New(repo domain.Repository, rewardClient domain.RewardClient) Usecase {
	return &service{repo: repo, rewardClient: rewardClient}
}

func (s *service) CreateRoom(ctx context.Context, name, creatorID string, limit int) (*domain.StudySession, error) {
	sess := &domain.StudySession{ID: "sess_test", Name: name, CreatorID: creatorID, ParticipantLimit: limit, Status: "ACTIVE"}
	return sess, s.repo.CreateSession(ctx, sess)
}

func (s *service) GetRoom(ctx context.Context, id string) (*domain.StudySession, error) {
	return s.repo.GetSession(ctx, id)
}

func (s *service) EndSession(ctx context.Context, sessionID, userID string) (*domain.StudySession, error) {
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		sess = &domain.StudySession{ID: sessionID, Name: "Ended Room", CreatorID: userID, Status: "ENDED"}
	} else {
		sess.Status = "ENDED"
	}

	// Trigger collaboration to Reward Service: AwardReward()
	if s.rewardClient != nil {
		_ = s.rewardClient.AwardReward(ctx, userID, "EndSession")
	}

	return sess, nil
}
