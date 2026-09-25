package usecase

import (
	"context"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

type Usecase interface {
	SignIn(ctx context.Context, email, name string) (*domain.UserAccount, error)
	SignUp(ctx context.Context, email, name string) (*domain.UserAccount, error)
	SignOut(ctx context.Context, userID string) error
	ViewProfile(ctx context.Context, userID string) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID, displayName string) (*domain.Profile, error)
	ViewStatistics(ctx context.Context, userID string) (*domain.UserStatistics, error)
}

type service struct {
	repo domain.Repository
}

func New(repo domain.Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) SignIn(ctx context.Context, email, name string) (*domain.UserAccount, error) {
	if email == "" {
		return nil, domain.ErrInvalid
	}
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return u, nil
	}
	// First login: auto sign up
	return s.SignUp(ctx, email, name)
}

func (s *service) SignUp(ctx context.Context, email, name string) (*domain.UserAccount, error) {
	if email == "" {
		return nil, domain.ErrInvalid
	}
	user := &domain.UserAccount{
		UserID:      "usr_" + time.Now().Format("20060102150405"),
		Email:       email,
		DisplayName: name,
		CreatedAt:   time.Now().UTC(),
	}
	return user, s.repo.CreateUser(ctx, user)
}

func (s *service) SignOut(ctx context.Context, userID string) error {
	return nil
}

func (s *service) ViewProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return s.repo.GetProfile(ctx, userID)
}

func (s *service) UpdateProfile(ctx context.Context, userID, displayName string) (*domain.Profile, error) {
	p, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	p.DisplayName = displayName
	p.UpdatedAt = time.Now().UTC()
	return p, s.repo.UpdateProfile(ctx, p)
}

func (s *service) ViewStatistics(ctx context.Context, userID string) (*domain.UserStatistics, error) {
	return &domain.UserStatistics{
		UserID:        userID,
		TotalSessions: 10,
		TotalFocusMin: 250,
		RewardsEarned: 5,
	}, nil
}
