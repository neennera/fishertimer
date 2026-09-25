// Package middleware verifies the account service's session JWT and forwards
// the caller's identity to downstream services as trusted headers. It is a
// read-only copy of the verification half of
// services/account/internal/adapter/token/jwt.go - only the standard library
// is used, so no cross-module import is needed (each service is its own Go
// module per go.work). Keep the payload shape and SessionCookie name in sync
// with the account service if either changes.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SessionCookie must match services/account/internal/adapter/handler.SessionCookie.
const SessionCookie = "ft_session"

// Headers the gateway sets from a verified token. Any of these arriving from
// the client is always stripped first, so downstream services can trust them.
const (
	HeaderUserID      = "X-User-Id"
	HeaderUserRole    = "X-User-Role"
	HeaderDisplayName = "X-Display-Name"
)

const useSession = "session"

var errInvalidToken = errors.New("middleware: invalid session token")

type header struct {
	Alg string `json:"alg"`
}

// payload mirrors the account service's session payload (a subset - Picture,
// used only by sign-up tickets, is intentionally omitted).
type payload struct {
	Sub  string `json:"sub,omitempty"`
	Role string `json:"role,omitempty"`
	Name string `json:"name,omitempty"`
	Use  string `json:"use"`
	Iss  string `json:"iss"`
	Exp  int64  `json:"exp"`
}

// Claims is the subset of the session JWT the gateway trusts and forwards.
type Claims struct {
	UserID      string
	Role        string
	DisplayName string
}

// Verifier checks session tokens signed by the account service.
type Verifier struct {
	secret []byte
	issuer string
	now    func() time.Time
}

func NewVerifier(secret, issuer string) *Verifier {
	return &Verifier{secret: []byte(secret), issuer: issuer, now: time.Now}
}

// Verify checks the algorithm, signature, issuer, expiry and that the token
// was issued as a session (not a sign-up ticket).
func (v *Verifier) Verify(raw string) (*Claims, error) {
	if len(v.secret) == 0 {
		return nil, errInvalidToken
	}

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	headBytes, err := decode(parts[0])
	if err != nil {
		return nil, errInvalidToken
	}
	var head header
	if err := json.Unmarshal(headBytes, &head); err != nil {
		return nil, errInvalidToken
	}
	if head.Alg != "HS256" {
		return nil, errInvalidToken
	}

	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := mac.Sum(nil)
	actual, err := decode(parts[2])
	if err != nil {
		return nil, errInvalidToken
	}
	if !hmac.Equal(expected, actual) {
		return nil, errInvalidToken
	}

	bodyBytes, err := decode(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}
	var body payload
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, errInvalidToken
	}

	if v.issuer != "" && body.Iss != v.issuer {
		return nil, errInvalidToken
	}
	if body.Exp == 0 || v.now().UTC().After(time.Unix(body.Exp, 0).UTC()) {
		return nil, errInvalidToken
	}
	if body.Use != useSession {
		return nil, errInvalidToken
	}

	return &Claims{UserID: body.Sub, Role: body.Role, DisplayName: body.Name}, nil
}

func decode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

// bearerOrCookie mirrors the account service's own bearerOrCookie: the
// browser session cookie, or an Authorization: Bearer header for
// service-to-service calls.
func bearerOrCookie(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}

// Identity always strips any client-supplied X-User-*/X-Display-Name headers
// first (so a caller cannot spoof identity), then - if a valid session token
// is present - re-sets them from the verified claims. It never rejects the
// request itself; enforcing that a route requires a session is left to each
// downstream service.
func (v *Verifier) Identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(HeaderUserID)
		r.Header.Del(HeaderUserRole)
		r.Header.Del(HeaderDisplayName)

		if raw := bearerOrCookie(r); raw != "" {
			if claims, err := v.Verify(raw); err == nil {
				r.Header.Set(HeaderUserID, claims.UserID)
				r.Header.Set(HeaderUserRole, claims.Role)
				r.Header.Set(HeaderDisplayName, url.QueryEscape(claims.DisplayName))
			}
		}

		next.ServeHTTP(w, r)
	})
}
