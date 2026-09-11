package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/account/internal/domain"
)

type Usecase interface {
	GetProfile(ctx context.Context, userID string) (*domain.Profile, error)
	SetBanStatus(ctx context.Context, userID string, banned bool) error
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return s.repo.GetProfile(ctx, userID)
}

func (s *service) SetBanStatus(ctx context.Context, userID string, banned bool) error {
	return s.repo.UpdateBanStatus(ctx, userID, banned)
}
