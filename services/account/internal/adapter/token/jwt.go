// Package token implements the session token required by UC-06: a compact JWS
// (JWT) signed with HMAC-SHA256, carrying user_id and role, valid for 7 days.
// Only the standard library is used, so any service in the monorepo can verify
// a token without pulling in a third-party dependency.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
)

var (
	ErrMalformed = errors.New("account: malformed token")
	ErrSignature = errors.New("account: invalid token signature")
	ErrExpired   = errors.New("account: token expired")
)

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// payload is the wire format. `sub` holds the user_id (standard JWT subject)
// and `role` is the authorisation claim other services read.
type payload struct {
	Sub   string `json:"sub"`
	Role  string `json:"role"`
	Email string `json:"email"`
	Iss   string `json:"iss"`
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
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
	if len(s.secret) == 0 {
		return "", nil, errors.New("account: JWT_SECRET is not configured")
	}

	issued := s.now().UTC()
	expires := issued.Add(s.ttl)

	head, err := json.Marshal(header{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", nil, err
	}
	body, err := json.Marshal(payload{
		Sub:   user.UserID,
		Role:  user.Role,
		Email: user.Email,
		Iss:   s.issuer,
		Iat:   issued.Unix(),
		Exp:   expires.Unix(),
	})
	if err != nil {
		return "", nil, err
	}

	signingInput := encode(head) + "." + encode(body)
	token := signingInput + "." + encode(s.sign(signingInput))

	return token, &domain.TokenClaims{
		UserID:    user.UserID,
		Role:      user.Role,
		Email:     user.Email,
		ExpiresAt: expires,
	}, nil
}

// Verify checks the signature, the algorithm, the issuer and the expiry.
func (s *JWTService) Verify(raw string) (*domain.TokenClaims, error) {
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

	return &domain.TokenClaims{
		UserID:    body.Sub,
		Role:      body.Role,
		Email:     body.Email,
		ExpiresAt: time.Unix(body.Exp, 0).UTC(),
	}, nil
}

func (s *JWTService) sign(input string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}

func encode(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func decode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
