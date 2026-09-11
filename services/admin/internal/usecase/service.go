package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
)

type Usecase interface {
	SubmitReport(ctx context.Context, reporterID, reportedID, reason string) (*domain.ModerationReport, error)
	GetPendingReports(ctx context.Context) ([]domain.ModerationReport, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) SubmitReport(ctx context.Context, reporterID, reportedID, reason string) (*domain.ModerationReport, error) {
	r := &domain.ModerationReport{ID: "rep_test", ReporterID: reporterID, ReportedID: reportedID, Reason: reason, Status: "PENDING"}
	return r, s.repo.CreateReport(ctx, r)
}

func (s *service) GetPendingReports(ctx context.Context) ([]domain.ModerationReport, error) {
	return s.repo.ListPending(ctx)
}
