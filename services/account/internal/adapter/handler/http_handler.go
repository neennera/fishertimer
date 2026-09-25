package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

// Cookies. ft_session is the credential the browser replays on every call;
// ft_signup holds a new user's verified Google identity while they pick a
// display name. The ft_oauth_* pair only lives for the few seconds of the
// Google round trip: the anti-CSRF state, and the two frontend paths the
// caller asked us to come back to.
const (
	SessionCookie    = "ft_session"
	SignUpCookie     = "ft_signup"
	StateCookie      = "ft_oauth_state"
	NextCookie       = "ft_oauth_next"
	SignUpPathCookie = "ft_oauth_signup_path"
)

// Where the browser is sent after Google when the caller did not say.
const (
	DefaultNextPath   = "/"
	DefaultSignUpPath = "/signup"
)

// The three states GET /me reports. The frontend decides what to show - and
// where to send the user - from this one value.
const (
	StatusSignedIn    = "signed_in"
	StatusNeedsSignUp = "needs_signup"
	StatusSignedOut   = "signed_out"
)

// Options carries the transport-level settings main.go derives from config.
type Options struct {
	FrontendURL     string
	CookieSecure    bool
	SessionMaxAge   time.Duration
	OAuthConfigured bool
}

type HTTPHandler struct {
	uc   usecase.Usecase
	opts Options
}

func New(uc usecase.Usecase, opts Options) *HTTPHandler {
	if opts.FrontendURL == "" {
		opts.FrontendURL = "http://localhost:3000"
	}
	if opts.SessionMaxAge == 0 {
		opts.SessionMaxAge = 7 * 24 * time.Hour
	}
	return &HTTPHandler{uc: uc, opts: opts}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/account/status", h.Status)

	// UC-06: Auth (SignIn & SignUp) via Google OAuth 2.0.
	mux.HandleFunc("/api/v1/account/google/login", h.GoogleLogin)
	mux.HandleFunc("/api/v1/account/google/callback", h.GoogleCallback)
	mux.HandleFunc("/api/v1/account/me", h.Me)
	mux.HandleFunc("/api/v1/account/signup", h.SignUp)
	mux.HandleFunc("/api/v1/account/signout", h.SignOut)

	// mux.HandleFunc("/api/v1/account/profile", h.Profile)
	// mux.HandleFunc("/api/v1/account/statistics", h.Statistics)
}

func (h *HTTPHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if !h.opts.OAuthConfigured {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "google oauth is not configured: set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET and GOOGLE_REDIRECT_URI",
		})
		return
	}

	state, err := randomState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "cannot start sign-in"})
		return
	}

	query := r.URL.Query()
	h.setCookie(w, StateCookie, state, 10*time.Minute)
	h.setCookie(w, NextCookie, safePath(query.Get("next"), DefaultNextPath), 10*time.Minute)
	h.setCookie(w, SignUpPathCookie, safePath(query.Get("signup"), DefaultSignUpPath), 10*time.Minute)

	http.Redirect(w, r, h.uc.SignInURL(state), http.StatusFound)
}

// GoogleCallback is where Google sends the browser back with ?code & ?state.
// It checks the state, completes the sign-in, and sends the browser to one of
// the two paths remembered at login:
//
//	e-mail has an account -> sign in, go to `next`
//	e-mail is new         -> no row written, go to the sign-up page with a
//	                         15-minute ticket cookie holding the identity
func (h *HTTPHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Where to land, as asked for at login. Read before the cookies are cleared.
	nextPath := safePath(cookieValue(r, NextCookie), DefaultNextPath)
	signUpPath := safePath(cookieValue(r, SignUpPathCookie), DefaultSignUpPath)
	h.clearCookie(w, StateCookie)
	h.clearCookie(w, NextCookie)
	h.clearCookie(w, SignUpPathCookie)

	if errParam := query.Get("error"); errParam != "" { // user pressed Cancel
		h.redirectTo(w, r, nextPath, errParam)
		return
	}

	// Constant-time compare so the check cannot be probed byte by byte.
	expected := cookieValue(r, StateCookie)
	if expected == "" || subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(expected)) != 1 {
		h.redirectTo(w, r, nextPath, "invalid_state")
		return
	}

	result, err := h.uc.CompleteSignIn(r.Context(), query.Get("code"))
	if err != nil {
		log.Printf("account: sign-in failed: %v", err)
		h.redirectTo(w, r, nextPath, "login_failed")
		return
	}

	// No account for this e-mail: nothing is written yet, the user has to pick
	// a display name first.
	if result.Session == nil {
		h.setCookie(w, SignUpCookie, result.SignUpToken, time.Until(result.SignUp.ExpiresAt))
		h.redirectTo(w, r, signUpPath, "")
		return
	}

	// Known e-mail: signed in.
	h.clearCookie(w, SignUpCookie)
	h.setCookie(w, SessionCookie, result.Session.Token, h.opts.SessionMaxAge)
	log.Printf("account: issued session token for %s: %s", result.Session.User.Email, result.Session.Token)
	h.redirectTo(w, r, nextPath, "")
}

// Me answers "who is this browser?" in one call, always 200:
//
//	{"status":"signed_in","user":{…}}       - has a valid session
//	{"status":"needs_signup","email":"…"}   - Google verified this e-mail and
//	                                          it has no account yet
//	{"status":"signed_out"}                 - neither
//
// The e-mail comes from the signed ft_signup cookie, never from the request,
// so this cannot be used to probe whether somebody else's address is
// registered.
func (h *HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	if user, err := h.uc.Authenticate(r.Context(), bearerOrCookie(r)); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": StatusSignedIn, "user": user})
		return
	}

	if ticket, err := h.uc.PendingSignUp(cookieValue(r, SignUpCookie)); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": StatusNeedsSignUp, "email": ticket.Email})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": StatusSignedOut})
}

// SignUp creates the account for the identity held in the ft_signup cookie and
// signs the user in.
//
//	POST {"display_name": "..."} -> 201 + the new user, sets ft_session
func (h *HTTPHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	session, err := h.uc.CompleteSignUp(r.Context(), cookieValue(r, SignUpCookie), req.DisplayName)
	switch {
	case errors.Is(err, domain.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": fmt.Sprintf("display_name is required (max %d characters)", domain.MaxDisplayNameLength),
		})
		return
	case errors.Is(err, domain.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "sign-up expired: sign in with Google again"})
		return
	case err != nil:
		log.Printf("account: sign-up failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "could not create account"})
		return
	}

	h.clearCookie(w, SignUpCookie)
	h.setCookie(w, SessionCookie, session.Token, h.opts.SessionMaxAge)
	log.Printf("account: issued session token for %s: %s", session.User.Email, session.Token)
	writeJSON(w, http.StatusCreated, session.User)
}

func (h *HTTPHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	h.clearCookie(w, SessionCookie)
	h.clearCookie(w, SignUpCookie)
	writeJSON(w, http.StatusOK, map[string]any{"signed_out": true})
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":   "account",
		"status":    "healthy",
		"port":      8082,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":          "account",
		"layer":            "adapter.handler",
		"oauth_configured": h.opts.OAuthConfigured,
		"ready":            true,
	})
}

// --- helpers ---------------------------------------------------------------

func (h *HTTPHandler) setCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,                 // JavaScript (and therefore XSS) cannot read it
		Secure:   h.opts.CookieSecure,  // HTTPS only in production
		SameSite: http.SameSiteLaxMode, // survives the redirect back from Google
		MaxAge:   int(ttl.Seconds()),
	})
}

func (h *HTTPHandler) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.opts.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// redirectTo sends the browser to a path on the frontend, optionally tagged
// with ?auth_error=<code> for the page to display.
func (h *HTTPHandler) redirectTo(w http.ResponseWriter, r *http.Request, path, errCode string) {
	target := strings.TrimSuffix(h.opts.FrontendURL, "/") + path
	if errCode != "" {
		target += "?auth_error=" + errCode
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// safePath keeps a caller-supplied redirect target on our own site. Anything
// that is not a plain absolute path - "//evil.example", "https://evil.example",
// a missing value - falls back to the default. Without this check, the
// ?next= parameter would be an open redirect.
func safePath(path, fallback string) string {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "\\") {
		return fallback
	}
	return path
}

// bearerOrCookie accepts the browser session cookie or an
// `Authorization: Bearer <jwt>` header (used by service-to-service calls).
func bearerOrCookie(r *http.Request) string {
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	return cookieValue(r, SessionCookie)
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
