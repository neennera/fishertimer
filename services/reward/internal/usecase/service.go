package usecase

import (
	"context"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

type Usecase interface {
	GrantReward(ctx context.Context, userID, species, rarity string) (*domain.FishReward, error)
	AwardReward(ctx context.Context, userID, reason string) (*domain.FishReward, error)
	GetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error)
	ListAllRewards(ctx context.Context) ([]domain.FishReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
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

func (s *service) AwardReward(ctx context.Context, userID, reason string) (*domain.FishReward, error) {
	r := &domain.FishReward{
		ID:      "fish_reward_" + userID,
		UserID:  userID,
		Species: "Golden Carp",
		Rarity:  "RARE",
	}
	return r, s.repo.Award(ctx, r)
}

func (s *service) GetUserInventory(ctx context.Context, userID string) ([]domain.FishReward, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) ListAllRewards(ctx context.Context) ([]domain.FishReward, error) {
	return s.repo.ListByUser(ctx, "")
}

func (s *service) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return s.repo.GetLastUpdate(ctx)
}

