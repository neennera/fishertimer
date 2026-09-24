package usecase_test

import (
	"context"
	"testing"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

type mockRepository struct{}

func (repo *mockRepository) GetRankings(ctx context.Context, period string) ([]domain.RankEntry, error) {
	return []domain.RankEntry{{UserID: "u1", DisplayName: "TopAngler", Rank: 1, RewardCount: 50, FocusMinutes: 1200, Period: period}}, nil
}

func (repo *mockRepository) UpsertScore(ctx context.Context, entry *domain.RankEntry) error {
	return nil
}

type mockRewardClient struct {
	called bool
	userID string
}

func (m *mockRewardClient) ViewRewards(ctx context.Context, userID string) ([]domain.FishReward, error) {
	m.called = true
	m.userID = userID
	return []domain.FishReward{{ID: "fish_1", UserID: userID, Species: "Golden Trout", Rarity: "LEGENDARY"}}, nil
}

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	rewardClient := &mockRewardClient{}
	svc := usecase.New(repo, rewardClient)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()

	ranks, err := svc.GetTopUsers(ctx, "weekly")
	if err != nil {
		t.Fatalf("unexpected error getting rankings: %v", err)
	}
	if len(ranks) != 1 {
		t.Fatalf("expected 1 rank entry, got %d", len(ranks))
	}

	rewards, err := svc.FetchUserRewards(ctx, "u1")
	if err != nil {
		t.Fatalf("unexpected error fetching rewards: %v", err)
	}
	if !rewardClient.called {
		t.Fatal("expected rewardClient.ViewRewards to be called")
	}
	if len(rewards) != 1 || rewards[0].Species != "Golden Trout" {
		t.Fatalf("expected 1 Golden Trout reward, got %v", rewards)
	}
}
