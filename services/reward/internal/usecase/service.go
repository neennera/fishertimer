package usecase

import (
	"context"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

type Usecase interface {
	AwardReward(ctx context.Context, userID, reason string) (*domain.UnlockedReward, error)

	GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) AwardReward(ctx context.Context, userID, reason string) (*domain.UnlockedReward, error) {
	r := &domain.UnlockedReward{
		UserID:   userID,
		ItemName: "Golden Carp",
		Category: domain.CategoryFish,
	}
	return r, s.repo.Award(ctx, r)
}

func (s *service) GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return s.repo.ListByUser(ctx, userID)
}
