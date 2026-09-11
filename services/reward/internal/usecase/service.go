package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

type Usecase interface {
	GrantReward(ctx context.Context, userID, species, rarity string) (*domain.FishReward, error)
	GetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) GrantReward(ctx context.Context, userID, species, rarity string) (*domain.FishReward, error) {
	r := &domain.FishReward{ID: "fish_test", UserID: userID, Species: species, Rarity: rarity}
	return r, s.repo.Award(ctx, r)
}

func (s *service) GetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error) {
	return s.repo.ListByUser(ctx, userID)
}
