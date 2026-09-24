package usecase

import (
	"context"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

type Usecase interface {
	GetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error)
	FetchUserRewards(ctx context.Context, userID string) ([]domain.FishReward, error)
}

type service struct {
	repo         domain.Repository
	rewardClient domain.RewardClient
}

func New(repo domain.Repository, rewardClient domain.RewardClient) Usecase {
	return &service{repo: repo, rewardClient: rewardClient}
}

func (s *service) GetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error) {
	return s.repo.GetRankings(ctx, period)
}

func (s *service) FetchUserRewards(ctx context.Context, userID string) ([]domain.FishReward, error) {
	if s.rewardClient == nil {
		return nil, nil
	}
	return s.rewardClient.ViewRewards(ctx, userID)
}
