package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sign(secret string, body payload) string {
	head, _ := json.Marshal(header{Alg: "HS256"})
	claims, _ := json.Marshal(body)
	signingInput := encode(head) + "." + encode(claims)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return signingInput + "." + encode(mac.Sum(nil))
}

func encode(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

const testSecret = "test-secret-test-secret-test-secret"
const testIssuer = "fishertimer-account"

func validToken() string {
	return sign(testSecret, payload{
		Sub:  "user-1",
		Role: "CUSTOMER",
		Name: "Student One",
		Use:  useSession,
		Iss:  testIssuer,
		Exp:  time.Now().Add(time.Hour).Unix(),
	})
}

func TestVerifier_Verify_ValidToken(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	claims, err := v.Verify(validToken())
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != "user-1" || claims.Role != "CUSTOMER" || claims.DisplayName != "Student One" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifier_Verify_TamperedSignature(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)
	raw := validToken()
	tampered := raw[:len(raw)-2] + "xx"

	if _, err := v.Verify(tampered); err == nil {
		t.Fatal("a tampered signature must be rejected")
	}
}

func TestVerifier_Verify_Expired(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)
	expired := sign(testSecret, payload{
		Sub: "user-1", Role: "CUSTOMER", Use: useSession, Iss: testIssuer,
		Exp: time.Now().Add(-time.Minute).Unix(),
	})

	if _, err := v.Verify(expired); err == nil {
		t.Fatal("an expired token must be rejected")
	}
}

func TestVerifier_Verify_WrongIssuerOrUse(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	wrongIssuer := sign(testSecret, payload{
		Sub: "user-1", Use: useSession, Iss: "someone-else", Exp: time.Now().Add(time.Hour).Unix(),
	})
	if _, err := v.Verify(wrongIssuer); err == nil {
		t.Fatal("a token from another issuer must be rejected")
	}

	signupTicket := sign(testSecret, payload{
		Use: "signup", Iss: testIssuer, Exp: time.Now().Add(time.Hour).Unix(),
	})
	if _, err := v.Verify(signupTicket); err == nil {
		t.Fatal("a sign-up ticket must not verify as a session")
	}
}

func TestVerifier_Verify_WrongSecret(t *testing.T) {
	v := NewVerifier("a-different-secret-a-different-secret", testIssuer)

	if _, err := v.Verify(validToken()); err == nil {
		t.Fatal("a token signed with another secret must be rejected")
	}
}

func TestIdentity_ValidToken_SetsHeaders(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	req.Header.Set("Authorization", "Bearer "+validToken())
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got.Header.Get(HeaderUserID) != "user-1" {
		t.Fatalf("expected X-User-Id=user-1, got %q", got.Header.Get(HeaderUserID))
	}
	if got.Header.Get(HeaderUserRole) != "CUSTOMER" {
		t.Fatalf("expected X-User-Role=CUSTOMER, got %q", got.Header.Get(HeaderUserRole))
	}
	if got.Header.Get(HeaderDisplayName) != "Student+One" {
		t.Fatalf("expected URL-escaped display name, got %q", got.Header.Get(HeaderDisplayName))
	}
}

func TestIdentity_ValidToken_ViaCookie(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookie, Value: validToken()})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got.Header.Get(HeaderUserID) != "user-1" {
		t.Fatalf("expected X-User-Id=user-1 via cookie, got %q", got.Header.Get(HeaderUserID))
	}
}

func TestIdentity_MissingToken_NoHeaders(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected passthrough 200, got %d", rec.Code)
	}
	if got.Header.Get(HeaderUserID) != "" || got.Header.Get(HeaderUserRole) != "" || got.Header.Get(HeaderDisplayName) != "" {
		t.Fatalf("expected no identity headers, got %+v", got.Header)
	}
}

func TestIdentity_InvalidToken_NoHeaders(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	req.Header.Set("Authorization", "Bearer garbage.not.a.token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got.Header.Get(HeaderUserID) != "" {
		t.Fatalf("expected no identity headers for an invalid token, got %+v", got.Header)
	}
}

func TestIdentity_SpoofedClientHeader_Stripped(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	req.Header.Set(HeaderUserID, "admin")
	req.Header.Set(HeaderUserRole, "ADMIN")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got.Header.Get(HeaderUserID) != "" || got.Header.Get(HeaderUserRole) != "" {
		t.Fatalf("client-supplied identity headers must be stripped, got %+v", got.Header)
	}
}

func TestIdentity_SpoofedClientHeader_OverwrittenByValidToken(t *testing.T) {
	v := NewVerifier(testSecret, testIssuer)

	var got *http.Request
	stub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r })
	handler := v.Identity(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/timer/status", nil)
	req.Header.Set(HeaderUserID, "admin")
	req.Header.Set("Authorization", "Bearer "+validToken())
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got.Header.Get(HeaderUserID) != "user-1" {
		t.Fatalf("expected the token's claims to win over the spoofed header, got %q", got.Header.Get(HeaderUserID))
	}
}
