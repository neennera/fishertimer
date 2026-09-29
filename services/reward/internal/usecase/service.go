package usecase

import (
	"context"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

type Usecase interface {
	AwardReward(ctx context.Context, userID, reason string) (*domain.UnlockedReward, error)
	GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error)
	ListAllRewards(ctx context.Context) ([]domain.UnlockedReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) AwardReward(ctx context.Context, userID, reason string) (*domain.UnlockedReward, error) {
	if userID == "" {
		return nil, domain.ErrInvalid
	}
	r := &domain.UnlockedReward{
		ItemID:   "reward_test",
		UserID:   userID,
		ItemName: "Golden Salmon",
		Category: domain.CategoryFish,
		Rarity:   domain.RarityRare,
	}
	return r, s.repo.Award(ctx, r)
}

func (s *service) GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) ListAllRewards(ctx context.Context) ([]domain.UnlockedReward, error) {
	return s.repo.ListByUser(ctx, "")
}

func (s *service) GetLastUpdate(ctx context.Context) (time.Time, error) {
	return s.repo.GetLastUpdate(ctx)
}
