package domain

import (
	"context"
)

type Repository interface {
	GetSession(ctx context.Context, id string) (*StudySession, error)
	CreateSession(ctx context.Context, s *StudySession) error
}
