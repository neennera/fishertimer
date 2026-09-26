package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// Usecase is the application layer. It orchestrates the OAuth provider, the
// user repository and the token service; it knows nothing about HTTP or SQL.
type Usecase interface {
	// SignInURL returns the Google consent-screen URL for this sign-in attempt.
	SignInURL(state string) string

	// CompleteSignIn turns Google's authorization code into either a session
	// (the e-mail already has an account: SignIn) or a sign-up ticket (it does
	// not: the user must pick a display name first). UC-06.
	CompleteSignIn(ctx context.Context, code string) (*domain.SignInResult, error)

	// PendingSignUp reads a sign-up ticket back so the form can show the
	// e-mail and suggest a display name.
	PendingSignUp(token string) (*domain.SignUpTicket, error)

	// CompleteSignUp creates the account from a sign-up ticket and the chosen
	// display name, then issues the session. UC-06 SignUp.
	CompleteSignUp(ctx context.Context, token, displayName string) (*domain.Session, error)

	// Authenticate verifies a session token and returns the live user record.
	Authenticate(ctx context.Context, token string) (*domain.UserAccount, error)

	// UpdateProfile renames the signed-in user and re-issues the session
	// token, since TokenClaims carries display_name for the gateway to forward
	// without a DB lookup.
	UpdateProfile(ctx context.Context, token, displayName string) (*domain.Session, error)

	// GetProfile looks up a user by id, for callers that already know which
	// account they want (e.g. another service resolving a user_id from a JWT).
	GetProfile(ctx context.Context, userID string) (*domain.UserAccount, error)
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

func (s *service) CompleteSignIn(ctx context.Context, code string) (*domain.SignInResult, error) {
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

	// UC-06 S-1: accounts are matched by e-mail. An existing user keeps their
	// stored role, so a pre-provisioned ADMIN stays ADMIN.
	email := strings.ToLower(strings.TrimSpace(profile.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		session, err := s.newSession(user)
		if err != nil {
			return nil, err
		}
		return &domain.SignInResult{Session: session}, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// No account yet: nothing is written until the user picks a display name.
	token, ticket, err := s.tokens.IssueSignUp(profile)
	if err != nil {
		return nil, err
	}
	return &domain.SignInResult{SignUpToken: token, SignUp: ticket}, nil
}

func (s *service) PendingSignUp(token string) (*domain.SignUpTicket, error) {
	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrUnauthorized
	}
	return s.tokens.VerifySignUp(token)
}

func (s *service) CompleteSignUp(ctx context.Context, token, displayName string) (*domain.Session, error) {
	ticket, err := s.PendingSignUp(token)
	if err != nil {
		return nil, err
	}

	name, err := validDisplayName(displayName)
	if err != nil {
		return nil, err
	}

	// The form can be submitted twice (double click, back button). If the
	// account already exists, sign into it rather than failing on the
	// UNIQUE(email) constraint.
	user, err := s.repo.GetUserByEmail(ctx, ticket.Email)
	if err == nil {
		return s.newSession(user)
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
		Email:       ticket.Email,
		DisplayName: name,
		AvatarURL:   ticket.Picture,
		Role:        domain.RoleCustomer, // UC-06 E-4: sign-up never produces ADMIN
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return s.newSession(user)
}

func (s *service) newSession(user *domain.UserAccount) (*domain.Session, error) {
	token, claims, err := s.tokens.Issue(user)
	if err != nil {
		return nil, err
	}
	return &domain.Session{Token: token, ExpiresAt: claims.ExpiresAt, User: user}, nil
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

func (s *service) UpdateProfile(ctx context.Context, token, displayName string) (*domain.Session, error) {
	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrUnauthorized
	}
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, err
	}

	name, err := validDisplayName(displayName)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.UpdateProfile(ctx, claims.UserID, name)
	if err != nil {
		return nil, err
	}
	return s.newSession(user)
}

func (s *service) GetProfile(ctx context.Context, userID string) (*domain.UserAccount, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrInvalid
	}
	return s.repo.GetUserByID(ctx, userID)
}

// validDisplayName trims the name and checks it fits users.display_name.
func validDisplayName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > domain.MaxDisplayNameLength {
		return "", domain.ErrInvalid
	}
	return name, nil
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
