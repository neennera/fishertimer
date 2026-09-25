package token_test

import (
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/token"
	"github.com/neennera/fishertimer/services/account/internal/domain"
)

func user() *domain.UserAccount {
	return &domain.UserAccount{
		UserID: "11111111-1111-4111-8111-111111111111",
		Email:  "student@example.com",
		Role:   domain.RoleCustomer,
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
	if claims.UserID != user().UserID || claims.Role != domain.RoleCustomer {
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
