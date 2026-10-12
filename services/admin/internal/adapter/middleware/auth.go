// Package middleware restricts the admin API to signed-in administrators
// (UC-04 E-1/E-5, UC-10 E-1/E-5). The admin website reaches this service
// directly, not through the API Gateway, so the account service's session
// JWT is verified here. Verification is a copy of
// services/api-gateway/internal/adapter/middleware/auth.go (standard library
// only; each service is its own Go module). Keep the payload shape and
// SessionCookie name in sync with the account service if either changes.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// SessionCookie must match services/account/internal/adapter/handler.SessionCookie.
const SessionCookie = "ft_session"

// RoleAdmin must match services/account/internal/domain.RoleAdmin.
const RoleAdmin = "ADMIN"

const useSession = "session"

var errInvalidToken = errors.New("middleware: invalid session token")

type header struct {
	Alg string `json:"alg"`
}

type payload struct {
	Sub  string `json:"sub,omitempty"`
	Role string `json:"role,omitempty"`
	Name string `json:"name,omitempty"`
	Use  string `json:"use"`
	Iss  string `json:"iss"`
	Exp  int64  `json:"exp"`
}

// Claims is the subset of the session JWT the admin service relies on.
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

// bearerOrCookie reads the browser session cookie, or an
// Authorization: Bearer header for service-to-service calls.
func bearerOrCookie(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}

type ctxKey struct{}

// RequireAdmin rejects a missing or expired session with 401, so the website
// sends the user to Sign In, and a non-admin session with 403, before any
// session data is read or any moderation command is sent.
func (v *Verifier) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearerOrCookie(r)
		if raw == "" {
			writeJSONError(w, http.StatusUnauthorized, "sign in required")
			return
		}
		claims, err := v.Verify(raw)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "session is invalid or expired")
			return
		}
		if claims.Role != RoleAdmin || claims.UserID == "" {
			writeJSONError(w, http.StatusForbidden, "admin role required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims)))
	})
}

// AdminFrom returns the verified admin set by RequireAdmin.
func AdminFrom(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
