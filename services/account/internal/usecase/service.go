package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// Usecase is the application layer. It orchestrates the OAuth provider, the
// user repository and the token service; it knows nothing about HTTP or SQL.
type Usecase interface {
	// SignInURL returns the Google consent-screen URL for this sign-in attempt.
	SignInURL(state string) string

	// CompleteSignIn turns Google's authorization code into a session:
	// read the profile -> match the account by e-mail (or create it on first
	// login) -> issue the JWT. This is UC-06 SignIn + SignUp.
	CompleteSignIn(ctx context.Context, code string) (*domain.Session, error)

	// Authenticate verifies a session token and returns the live user record.
	Authenticate(ctx context.Context, token string) (*domain.UserAccount, error)

	ViewProfile(ctx context.Context, userID string) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, userID, displayName string) (*domain.Profile, error)
	ViewStatistics(ctx context.Context, userID string) (*domain.UserStatistics, error)
}

type service struct {
	repo     domain.Repository
	provider domain.OAuthProvider
	tokens   domain.TokenService
}

func New(repo domain.Repository, provider domain.OAuthProvider, tokens domain.TokenService) Usecase {
	return &service{repo: repo, provider: provider, tokens: tokens}
}

func (s *service) SignInURL(state string) string {
	return s.provider.AuthCodeURL(state)
}

func (s *service) CompleteSignIn(ctx context.Context, code string) (*domain.Session, error) {
	if strings.TrimSpace(code) == "" {
		return nil, domain.ErrInvalid
	}

	profile, err := s.provider.FetchProfile(ctx, code)
	if err != nil {
		return nil, err
	}
	if !profile.Valid() {
		return nil, domain.ErrInvalid
	}

	user, err := s.matchOrCreateUser(ctx, profile)
	if err != nil {
		return nil, err
	}

	token, claims, err := s.tokens.Issue(user)
	if err != nil {
		return nil, err
	}
	return &domain.Session{Token: token, ExpiresAt: claims.ExpiresAt, User: user}, nil
}

// matchOrCreateUser is UC-06 S-1: match by e-mail, otherwise create a new
// account with role CUSTOMER. An existing user keeps their stored role, so a
// pre-provisioned ADMIN stays ADMIN.
func (s *service) matchOrCreateUser(ctx context.Context, profile *domain.GoogleProfile) (*domain.UserAccount, error) {
	email := strings.ToLower(strings.TrimSpace(profile.Email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	userID, err := newUserID()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	user = &domain.UserAccount{
		UserID:      userID,
		Email:       email,
		DisplayName: displayName(profile, email),
		AvatarURL:   profile.Picture,
		Role:        domain.RoleCustomer,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) Authenticate(ctx context.Context, token string) (*domain.UserAccount, error) {
	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrUnauthorized
	}
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(ctx, claims.UserID)
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
	// Placeholder until the Timer and Reward collaborations are implemented.
	return &domain.UserStatistics{
		UserID:        userID,
		TotalSessions: 10,
		TotalFocusMin: 250,
		RewardsEarned: 5,
	}, nil
}

// displayName falls back to the local part of the e-mail: the column is NOT NULL.
func displayName(profile *domain.GoogleProfile, email string) string {
	if name := strings.TrimSpace(profile.Name); name != "" {
		return name
	}
	if local, _, found := strings.Cut(email, "@"); found && local != "" {
		return local
	}
	return "Fisher"
}

// newUserID returns a version 4 UUID, matching the `user_id uuid` column.
func newUserID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32]), nil
}
