package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

type stubProvider struct{ profile *domain.GoogleProfile }

func (p *stubProvider) AuthCodeURL(state string) string {
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
}

func (p *stubProvider) FetchProfile(ctx context.Context, code string) (*domain.GoogleProfile, error) {
	return p.profile, nil
}

type stubTokens struct{}

func (t *stubTokens) Issue(u *domain.UserAccount) (string, *domain.TokenClaims, error) {
	return "token." + u.UserID, &domain.TokenClaims{
		UserID:    u.UserID,
		Role:      u.Role,
		Email:     u.Email,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

func (t *stubTokens) Verify(token string) (*domain.TokenClaims, error) {
	if len(token) < 7 || token[:6] != "token." {
		return nil, domain.ErrUnauthorized
	}
	return &domain.TokenClaims{UserID: token[6:]}, nil
}

func newService(profile *domain.GoogleProfile) (usecase.Usecase, *repository.InMemoryRepository) {
	repo := repository.NewInMemory()
	return usecase.New(repo, &stubProvider{profile: profile}, &stubTokens{}), repo
}

func googleProfile() *domain.GoogleProfile {
	return &domain.GoogleProfile{
		Email:         "Student@Example.com",
		EmailVerified: true,
		Name:          "Student One",
		Picture:       "https://lh3.googleusercontent.com/avatar",
	}
}

func TestCompleteSignIn_CreatesAccountOnFirstLogin(t *testing.T) {
	svc, repo := newService(googleProfile())

	session, err := svc.CompleteSignIn(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if session.Token == "" {
		t.Fatal("expected a session token")
	}
	if session.User.Email != "student@example.com" {
		t.Fatalf("e-mail should be lower-cased, got %q", session.User.Email)
	}
	if session.User.Role != domain.RoleCustomer {
		t.Fatalf("a new account must be CUSTOMER, got %q", session.User.Role)
	}
	if len(session.User.UserID) != 36 {
		t.Fatalf("user_id should be a uuid, got %q", session.User.UserID)
	}

	if _, err := repo.GetUserByEmail(context.Background(), "student@example.com"); err != nil {
		t.Fatalf("account was not persisted: %v", err)
	}
}

func TestCompleteSignIn_ReusesAccountOnSecondLogin(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	first, err := svc.CompleteSignIn(ctx, "code-1")
	if err != nil {
		t.Fatalf("first sign-in: %v", err)
	}
	second, err := svc.CompleteSignIn(ctx, "code-2")
	if err != nil {
		t.Fatalf("second sign-in: %v", err)
	}

	if first.User.UserID != second.User.UserID {
		t.Fatalf("expected the same account, got %q then %q", first.User.UserID, second.User.UserID)
	}
}

func TestCompleteSignIn_KeepsAdminRole(t *testing.T) {
	svc, repo := newService(googleProfile())
	ctx := context.Background()

	// A pre-provisioned ADMIN row (UC-06 E-4).
	admin := &domain.UserAccount{
		UserID:      "11111111-1111-4111-8111-111111111111",
		Email:       "student@example.com",
		DisplayName: "Admin",
		Role:        domain.RoleAdmin,
	}
	if err := repo.CreateUser(ctx, admin); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	session, err := svc.CompleteSignIn(ctx, "code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if session.User.Role != domain.RoleAdmin {
		t.Fatalf("an existing ADMIN must keep their role, got %q", session.User.Role)
	}
}

func TestCompleteSignIn_RejectsUnverifiedEmail(t *testing.T) {
	profile := googleProfile()
	profile.EmailVerified = false
	svc, _ := newService(profile)

	if _, err := svc.CompleteSignIn(context.Background(), "code"); err == nil {
		t.Fatal("an unverified google e-mail must be rejected")
	}
}

func TestAuthenticate(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	session, err := svc.CompleteSignIn(ctx, "code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}

	user, err := svc.Authenticate(ctx, session.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.UserID != session.User.UserID {
		t.Fatal("authenticated user does not match the session")
	}

	if _, err := svc.Authenticate(ctx, "not-a-token"); err == nil {
		t.Fatal("an invalid token must be rejected")
	}
}
