package usecase

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
)

// AwardInput is one completed work cycle to award.
type AwardInput struct {
	UserID           string
	SessionID        string
	CycleID          string
	WorkMinutes      int
	ParticipantCount int
	DisplayName      string // optional; stored on each user_rewards row
}

// AwardResult is the rewards for a cycle. AlreadyAwarded is true when they
// were stored by an earlier request and nothing new was inserted.
type AwardResult struct {
	CycleID        string
	AlreadyAwarded bool
	Rewards        []domain.UnlockedReward
}

type Usecase interface {
	AwardReward(ctx context.Context, in AwardInput) (*AwardResult, error)
	GetUserInventory(ctx context.Context, userID string) ([]domain.UnlockedReward, error)
	ListAllRewards(ctx context.Context) ([]domain.UnlockedReward, error)
	GetLastUpdate(ctx context.Context) (time.Time, error)
}

type service struct {
	repo domain.Repository
	rnd  func() float64
}

func New(repo domain.Repository) Usecase {
	return NewWithRand(repo, rand.Float64)
}

// NewWithRand takes the random source for reward draws (values in [0, 1)),
// so tests can make draws deterministic.
func NewWithRand(repo domain.Repository, rnd func() float64) Usecase {
	return &service{repo: repo, rnd: rnd}
}

// AwardReward draws and stores the rewards for one completed work cycle. It
// is idempotent per cycle_id: a retry returns the stored rewards (E-1).
func (s *service) AwardReward(ctx context.Context, in AwardInput) (*AwardResult, error) {
	if in.UserID == "" || in.CycleID == "" || in.WorkMinutes < 0 {
		return nil, domain.ErrInvalid
	}

	existing, err := s.repo.ListByCycle(ctx, in.CycleID)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return &AwardResult{CycleID: in.CycleID, AlreadyAwarded: true, Rewards: existing}, nil
	}

	// Under 15 minutes earns nothing and stores nothing (E-3).
	if domain.RewardCount(in.WorkMinutes) == 0 {
		return &AwardResult{CycleID: in.CycleID, Rewards: []domain.UnlockedReward{}}, nil
	}

	catalogue, err := s.repo.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	drawn, err := domain.DrawRewards(catalogue, in.WorkMinutes, in.ParticipantCount, s.rnd)
	if err != nil {
		return nil, err
	}

	awardedAt := time.Now().UTC()
	rewards := make([]domain.UnlockedReward, len(drawn))
	for i, item := range drawn {
		rewards[i] = domain.UnlockedReward{
			ItemID:      item.ID,
			UserID:      in.UserID,
			DisplayName: in.DisplayName,
			CycleID:     in.CycleID,
			ItemName:    item.ItemName,
			Species:     item.ItemName,
			Category:    item.Category,
			Rarity:      item.Rarity,
			BaseWeight:  item.BaseWeight,
			ScoreValue:  item.ScoreValue,
			AssetURL:    item.AssetURL,
			AwardedAt:   awardedAt,
		}
	}

	if err := s.repo.AwardMany(ctx, rewards); err != nil {
		if !errors.Is(err, domain.ErrAlreadyAwarded) {
			return nil, err
		}
		// A concurrent retry stored this cycle first: return its rewards.
		existing, err := s.repo.ListByCycle(ctx, in.CycleID)
		if err != nil {
			return nil, err
		}
		return &AwardResult{CycleID: in.CycleID, AlreadyAwarded: true, Rewards: existing}, nil
	}
	return &AwardResult{CycleID: in.CycleID, Rewards: rewards}, nil
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
