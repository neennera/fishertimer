// Package token implements the session token required by UC-06: a compact JWS
// (JWT) signed with HMAC-SHA256, carrying user_id and role, valid for 7 days.
// It also signs the short-lived sign-up ticket that carries a new user's
// verified Google identity to the display-name form.
// Only the standard library is used, so any service in the monorepo can verify
// a token without pulling in a third-party dependency.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

// Every rejection wraps domain.ErrUnauthorized, so callers outside this
// package only need errors.Is(err, domain.ErrUnauthorized).
var (
	ErrMalformed = fmt.Errorf("%w: malformed token", domain.ErrUnauthorized)
	ErrSignature = fmt.Errorf("%w: invalid token signature", domain.ErrUnauthorized)
	ErrExpired   = fmt.Errorf("%w: token expired", domain.ErrUnauthorized)
	ErrWrongUse  = fmt.Errorf("%w: token used for the wrong purpose", domain.ErrUnauthorized)
)

// SignUpTTL is how long a new user has to pick a display name.
const SignUpTTL = 15 * time.Minute

// Values of the `use` claim. Both kinds share one secret, so the claim is what
// stops a sign-up ticket from being replayed as a session and vice versa.
const (
	useSession = "session"
	useSignUp  = "signup"
)

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// payload is the wire format. `sub` holds the user_id (standard JWT subject)
// and `role` is the authorisation claim other services read. `picture` is only
// set on sign-up tickets, where it becomes the new row's avatar_url.
type payload struct {
	Sub     string `json:"sub,omitempty"`
	Role    string `json:"role,omitempty"`
	Email   string `json:"email"`
	Picture string `json:"picture,omitempty"`
	Use     string `json:"use"`
	Iss     string `json:"iss"`
	Iat     int64  `json:"iat"`
	Exp     int64  `json:"exp"`
}

type JWTService struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func New(secret, issuer string, ttl time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), issuer: issuer, ttl: ttl, now: time.Now}
}

// Issue signs a session token for the given account.
func (s *JWTService) Issue(user *domain.UserAccount) (string, *domain.TokenClaims, error) {
	if user == nil || user.UserID == "" {
		return "", nil, domain.ErrInvalid
	}

	issued := s.now().UTC()
	expires := issued.Add(s.ttl)

	token, err := s.seal(payload{
		Sub:   user.UserID,
		Role:  user.Role,
		Email: user.Email,
		Use:   useSession,
		Iat:   issued.Unix(),
		Exp:   expires.Unix(),
	})
	if err != nil {
		return "", nil, err
	}

	return token, &domain.TokenClaims{
		UserID:    user.UserID,
		Role:      user.Role,
		Email:     user.Email,
		ExpiresAt: expires,
	}, nil
}

// Verify checks a session token: signature, algorithm, issuer, expiry and use.
func (s *JWTService) Verify(raw string) (*domain.TokenClaims, error) {
	body, err := s.open(raw, useSession)
	if err != nil {
		return nil, err
	}

	return &domain.TokenClaims{
		UserID:    body.Sub,
		Role:      body.Role,
		Email:     body.Email,
		ExpiresAt: time.Unix(body.Exp, 0).UTC(),
	}, nil
}

// IssueSignUp signs a ticket for a verified Google identity with no account.
func (s *JWTService) IssueSignUp(profile *domain.GoogleProfile) (string, *domain.SignUpTicket, error) {
	if !profile.Valid() {
		return "", nil, domain.ErrInvalid
	}

	issued := s.now().UTC()
	expires := issued.Add(SignUpTTL)
	email := strings.ToLower(strings.TrimSpace(profile.Email))

	token, err := s.seal(payload{
		Email:   email,
		Picture: profile.Picture,
		Use:     useSignUp,
		Iat:     issued.Unix(),
		Exp:     expires.Unix(),
	})
	if err != nil {
		return "", nil, err
	}

	return token, &domain.SignUpTicket{
		Email:     email,
		Picture:   profile.Picture,
		ExpiresAt: expires,
	}, nil
}

// VerifySignUp checks a sign-up ticket the same way Verify checks a session.
func (s *JWTService) VerifySignUp(raw string) (*domain.SignUpTicket, error) {
	body, err := s.open(raw, useSignUp)
	if err != nil {
		return nil, err
	}
	if body.Email == "" {
		return nil, ErrMalformed
	}

	return &domain.SignUpTicket{
		Email:     body.Email,
		Picture:   body.Picture,
		ExpiresAt: time.Unix(body.Exp, 0).UTC(),
	}, nil
}

// seal stamps the issuer and signs the payload as header.payload.signature.
func (s *JWTService) seal(body payload) (string, error) {
	if len(s.secret) == 0 {
		return "", errors.New("account: JWT_SECRET is not configured")
	}
	body.Iss = s.issuer

	head, err := json.Marshal(header{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	signingInput := encode(head) + "." + encode(claims)
	return signingInput + "." + encode(s.sign(signingInput)), nil
}

// open checks the algorithm, the signature, the issuer, the expiry and that
// the token was issued for `use`, then returns its payload.
func (s *JWTService) open(raw, use string) (*payload, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, ErrMalformed
	}

	headBytes, err := decode(parts[0])
	if err != nil {
		return nil, ErrMalformed
	}
	var head header
	if err := json.Unmarshal(headBytes, &head); err != nil {
		return nil, ErrMalformed
	}
	// Reject "alg":"none" and any algorithm-substitution attempt.
	if head.Alg != "HS256" {
		return nil, ErrSignature
	}

	expected := s.sign(parts[0] + "." + parts[1])
	actual, err := decode(parts[2])
	if err != nil {
		return nil, ErrMalformed
	}
	// Constant-time compare: never leak signature bytes through timing.
	if !hmac.Equal(expected, actual) {
		return nil, ErrSignature
	}

	bodyBytes, err := decode(parts[1])
	if err != nil {
		return nil, ErrMalformed
	}
	var body payload
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, ErrMalformed
	}

	if s.issuer != "" && body.Iss != s.issuer {
		return nil, ErrSignature
	}
	if body.Exp == 0 || s.now().UTC().After(time.Unix(body.Exp, 0).UTC()) {
		return nil, ErrExpired
	}
	if body.Use != use {
		return nil, ErrWrongUse
	}

	return &body, nil
}

func (s *JWTService) sign(input string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}

func encode(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func decode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
