package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

type Usecase interface {
	GetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) GetTopUsers(ctx context.Context, period string) ([]domain.RankEntry, error) {
	return s.repo.GetRankings(ctx, period)
}
