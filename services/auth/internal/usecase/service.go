package usecase

import (
	"context"
	"github.com/neennera/fishertimer/services/auth/internal/domain"
)

type Usecase interface {
	Authenticate(ctx context.Context, email, name string) (*domain.UserAccount, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) Authenticate(ctx context.Context, email, name string) (*domain.UserAccount, error) {
	if email == "" {
		return nil, domain.ErrInvalid
	}
	user := &domain.UserAccount{ID: "usr_test", Email: email, Name: name}
	return user, s.repo.Create(ctx, user)
}
