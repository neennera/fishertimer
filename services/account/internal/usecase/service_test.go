package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/account/internal/adapter/token"
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

func newService(profile *domain.GoogleProfile) (usecase.Usecase, *repository.InMemoryRepository) {
	repo := repository.NewInMemory()
	tokens := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	return usecase.New(repo, &stubProvider{profile: profile}, tokens), repo
}

func googleProfile() *domain.GoogleProfile {
	return &domain.GoogleProfile{
		Email:         "Student@Example.com",
		EmailVerified: true,
		Name:          "Student One",
		Picture:       "https://lh3.googleusercontent.com/avatar",
	}
}

// signUp runs the whole first-login flow: Google callback, then the form.
func signUp(t *testing.T, svc usecase.Usecase, displayName string) *domain.Session {
	t.Helper()
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if result.Session != nil || result.SignUpToken == "" {
		t.Fatalf("a new e-mail must get a sign-up ticket, got %+v", result)
	}

	session, err := svc.CompleteSignUp(ctx, result.SignUpToken, displayName)
	if err != nil {
		t.Fatalf("CompleteSignUp: %v", err)
	}
	return session
}

func TestCompleteSignIn_NewEmailDoesNotCreateAccountYet(t *testing.T) {
	svc, repo := newService(googleProfile())

	result, err := svc.CompleteSignIn(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if result.SignUpToken == "" || result.Session != nil {
		t.Fatalf("expected a sign-up ticket only, got %+v", result)
	}
	if _, err := repo.GetUserByEmail(context.Background(), "student@example.com"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no account may exist before the display name is chosen, got %v", err)
	}
}

func TestPendingSignUp_ReturnsVerifiedEmail(t *testing.T) {
	svc, _ := newService(googleProfile())

	result, err := svc.CompleteSignIn(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	ticket, err := svc.PendingSignUp(result.SignUpToken)
	if err != nil {
		t.Fatalf("PendingSignUp: %v", err)
	}
	// The e-mail comes from the signed ticket, not from the caller: that is
	// what makes it safe for the frontend to ask "does this need an account?".
	if ticket.Email != "student@example.com" {
		t.Fatalf("unexpected ticket: %+v", ticket)
	}
}

func TestCompleteSignUp_CreatesAccountWithChosenName(t *testing.T) {
	svc, repo := newService(googleProfile())

	session := signUp(t, svc, "  Fish Lover  ")

	if session.Token == "" {
		t.Fatal("expected a session token")
	}
	if session.User.DisplayName != "Fish Lover" {
		t.Fatalf("display name should be the trimmed input, got %q", session.User.DisplayName)
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

func TestCompleteSignUp_RejectsBadDisplayName(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}

	for _, name := range []string{"", "   ", strings.Repeat("a", domain.MaxDisplayNameLength+1)} {
		if _, err := svc.CompleteSignUp(ctx, result.SignUpToken, name); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("display name %q: expected ErrInvalid, got %v", name, err)
		}
	}
}

func TestCompleteSignUp_RejectsForgedOrSessionToken(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	if _, err := svc.CompleteSignUp(ctx, "not-a-token", "Name"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("forged ticket: expected ErrUnauthorized, got %v", err)
	}

	session := signUp(t, svc, "Name")
	if _, err := svc.CompleteSignUp(ctx, session.Token, "Other"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("a session token must not work as a sign-up ticket, got %v", err)
	}
}

func TestCompleteSignUp_SubmittedTwiceSignsIntoSameAccount(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	first, err := svc.CompleteSignUp(ctx, result.SignUpToken, "First")
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	second, err := svc.CompleteSignUp(ctx, result.SignUpToken, "Second")
	if err != nil {
		t.Fatalf("second submit: %v", err)
	}
	if first.User.UserID != second.User.UserID || second.User.DisplayName != "First" {
		t.Fatalf("expected the original account, got %+v", second.User)
	}
}

func TestCompleteSignIn_ExistingEmailSignsIn(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	created := signUp(t, svc, "Fish Lover")

	result, err := svc.CompleteSignIn(ctx, "code-2")
	if err != nil {
		t.Fatalf("second sign-in: %v", err)
	}
	if result.Session == nil || result.SignUpToken != "" {
		t.Fatalf("a known e-mail must sign straight in, got %+v", result)
	}
	if result.Session.User.UserID != created.User.UserID {
		t.Fatalf("expected the same account, got %q then %q", created.User.UserID, result.Session.User.UserID)
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

	result, err := svc.CompleteSignIn(ctx, "code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if result.Session == nil || result.Session.User.Role != domain.RoleAdmin {
		t.Fatalf("an existing ADMIN must sign in and keep their role, got %+v", result)
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

	session := signUp(t, svc, "Fish Lover")

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

func TestAuthenticate_RejectsSignUpTicket(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if _, err := svc.Authenticate(ctx, result.SignUpToken); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("a sign-up ticket must not work as a session, got %v", err)
	}
}
