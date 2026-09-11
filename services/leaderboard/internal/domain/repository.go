package domain

import (
	"context"
)

type Repository interface {
	GetRankings(ctx context.Context, period string) ([]RankEntry, error)
	UpsertScore(ctx context.Context, entry *RankEntry) error
}
