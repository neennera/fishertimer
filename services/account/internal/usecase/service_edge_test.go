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

// racingRepo plays the loser of a first-sign-up race: the other request's
// insert lands between our e-mail lookup and our insert, so CreateUser fails
// with a UNIQUE(email) violation although the account now exists.
type racingRepo struct {
	*repository.InMemoryRepository
	lookups int
}

func (r *racingRepo) GetUserByEmail(ctx context.Context, email string) (*domain.UserAccount, error) {
	r.lookups++
	return r.InMemoryRepository.GetUserByEmail(ctx, email)
}

func (r *racingRepo) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	winner := *u
	winner.UserID = "11111111-1111-4111-8111-111111111111"
	winner.DisplayName = "Winner"
	if err := r.InMemoryRepository.CreateUser(ctx, &winner); err != nil {
		return err
	}
	return errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`)
}

func TestCompleteSignUp_ConcurrentSubmitSignsIntoWinningAccount(t *testing.T) {
	repo := &racingRepo{InMemoryRepository: repository.NewInMemory()}
	tokens := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	svc := usecase.New(repo, &stubProvider{profile: googleProfile()}, tokens, &stubTimerClient{}, &stubRewardClient{})
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	session, err := svc.CompleteSignUp(ctx, result.SignUpToken, "Loser")
	if err != nil {
		t.Fatalf("a lost insert race must sign into the existing account, got %v", err)
	}
	if session.User.DisplayName != "Winner" {
		t.Fatalf("expected the winning account, got %+v", session.User)
	}
}

func TestCompleteSignUp_InsertFailureWithoutAccountStillErrors(t *testing.T) {
	// A genuine failure (no account appears) must not be swallowed.
	repo := &failingCreateRepo{InMemoryRepository: repository.NewInMemory()}
	tokens := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	svc := usecase.New(repo, &stubProvider{profile: googleProfile()}, tokens, &stubTimerClient{}, &stubRewardClient{})
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}
	if _, err := svc.CompleteSignUp(ctx, result.SignUpToken, "Name"); err == nil {
		t.Fatal("expected the insert error to be returned")
	}
}

type failingCreateRepo struct{ *repository.InMemoryRepository }

func (r *failingCreateRepo) CreateUser(ctx context.Context, u *domain.UserAccount) error {
	return errors.New("db down")
}

func TestDisplayName_RejectsInvisibleAndControlOnlyNames(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	result, err := svc.CompleteSignIn(ctx, "auth-code")
	if err != nil {
		t.Fatalf("CompleteSignIn: %v", err)
	}

	bad := []string{
		"\u200b\u200b",       // zero-width spaces only
		"\ufeff",             // BOM only
		" \u2060 \u200d ",    // word joiner / joiner only
		"Fish\x00Lover",      // control character inside
		"Fish\nLover",        // newline inside
		"\u200b\u200c\u200d", // every zero-width character
	}
	for _, name := range bad {
		if _, err := svc.CompleteSignUp(ctx, result.SignUpToken, name); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("display name %q: expected ErrInvalid, got %v", name, err)
		}
	}

	// Ordinary unicode (emoji, Thai) is fine and counted by character.
	for _, name := range []string{"ปลา 🐟", "Fish\u200bLover", strings.Repeat("a", domain.MaxDisplayNameLength), strings.Repeat("🐟", domain.MaxDisplayNameLength)} {
		if _, err := svc.CompleteSignUp(ctx, result.SignUpToken, name); err != nil {
			t.Fatalf("display name %q should be accepted, got %v", name, err)
		}
	}
}

func TestLookups_MalformedIDIsNotFoundNotAServerError(t *testing.T) {
	svc, _ := newService(googleProfile())
	ctx := context.Background()

	for _, id := range []string{"abc", "not-a-uuid", "1234", "00000000-0000-0000-0000-00000000000z"} {
		if _, err := svc.GetProfile(ctx, id); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("GetProfile(%q): expected ErrNotFound, got %v", id, err)
		}
		if _, err := svc.GetTimerStatistics(ctx, id); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("GetTimerStatistics(%q): expected ErrNotFound, got %v", id, err)
		}
		if _, err := svc.GetRewards(ctx, id); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("GetRewards(%q): expected ErrNotFound, got %v", id, err)
		}
	}
	if _, err := svc.GetProfile(ctx, "  "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("blank id: expected ErrInvalid, got %v", err)
	}
}
