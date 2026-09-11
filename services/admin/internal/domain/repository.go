package domain

import (
	"context"
)

type Repository interface {
	CreateReport(ctx context.Context, r *ModerationReport) error
	ListPending(ctx context.Context) ([]ModerationReport, error)
}
