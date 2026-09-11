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

func TestUsecase_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := usecase.New(repo)
	if svc == nil {
		t.Fatal("expected usecase service to be initialized")
	}
	ctx := context.Background()
	_ = ctx
}
