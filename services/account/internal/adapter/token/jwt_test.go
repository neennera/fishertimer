package token_test

import (
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/token"
	"github.com/neennera/fishertimer/services/account/internal/domain"
)

func user() *domain.UserAccount {
	return &domain.UserAccount{
		UserID:      "11111111-1111-4111-8111-111111111111",
		Email:       "student@example.com",
		DisplayName: "Student One",
		Role:        domain.RoleCustomer,
	}
}

func TestIssueAndVerify(t *testing.T) {
	svc := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)

	raw, _, err := svc.Issue(user())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	claims, err := svc.Verify(raw)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != user().UserID || claims.Role != domain.RoleCustomer || claims.DisplayName != user().DisplayName {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerify_RejectsOtherSecretAndExpiry(t *testing.T) {
	issuer := token.New("secret-one-secret-one-secret-one-ok", "fishertimer-account", time.Hour)
	attacker := token.New("secret-two-secret-two-secret-two-ok", "fishertimer-account", time.Hour)

	raw, _, err := issuer.Issue(user())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := attacker.Verify(raw); err == nil {
		t.Fatal("a token signed with another secret must be rejected")
	}

	expiredSvc := token.New("secret-one-secret-one-secret-one-ok", "fishertimer-account", -time.Minute)
	expired, _, err := expiredSvc.Issue(user())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := expiredSvc.Verify(expired); err != token.ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestVerify_RejectsAlgNone(t *testing.T) {
	svc := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)

	// {"alg":"none","typ":"JWT"} with an unsigned body.
	forged := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c3ItMSJ9."
	if _, err := svc.Verify(forged); err == nil {
		t.Fatal("alg=none must be rejected")
	}
}

func TestSignUpTicket_RoundTripAndSeparateFromSession(t *testing.T) {
	svc := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	profile := &domain.GoogleProfile{Email: "New@Example.com", EmailVerified: true, Name: "New Person"}

	raw, _, err := svc.IssueSignUp(profile)
	if err != nil {
		t.Fatalf("IssueSignUp: %v", err)
	}
	ticket, err := svc.VerifySignUp(raw)
	if err != nil {
		t.Fatalf("VerifySignUp: %v", err)
	}
	if ticket.Email != "new@example.com" {
		t.Fatalf("unexpected ticket: %+v", ticket)
	}

	if _, err := svc.Verify(raw); err != token.ErrWrongUse {
		t.Fatalf("a sign-up ticket must not verify as a session, got %v", err)
	}
	session, _, err := svc.Issue(user())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := svc.VerifySignUp(session); err != token.ErrWrongUse {
		t.Fatalf("a session must not verify as a sign-up ticket, got %v", err)
	}
}

func TestIssueSignUp_RejectsUnverifiedEmail(t *testing.T) {
	svc := token.New("test-secret-test-secret-test-secret", "fishertimer-account", time.Hour)
	if _, _, err := svc.IssueSignUp(&domain.GoogleProfile{Email: "x@example.com"}); err == nil {
		t.Fatal("an unverified e-mail must not get a sign-up ticket")
	}
}
