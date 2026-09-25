package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

// Cookies. ft_session is the credential the browser replays on every call;
// ft_oauth_state only lives for the few seconds of the Google round trip.
const (
	SessionCookie = "ft_session"
	StateCookie   = "ft_oauth_state"
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
	mux.HandleFunc("/api/v1/account/signout", h.SignOut)

	mux.HandleFunc("/api/v1/account/profile", h.Profile)
	mux.HandleFunc("/api/v1/account/statistics", h.Statistics)
}

// GoogleLogin starts the flow: mint an anti-CSRF state, park it in a
// short-lived HttpOnly cookie, and redirect to Google's consent screen.
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

	h.setCookie(w, StateCookie, state, 10*time.Minute)
	http.Redirect(w, r, h.uc.SignInURL(state), http.StatusFound)
}

// GoogleCallback is where Google sends the browser back with ?code & ?state.
// It checks the state, completes the sign-in and sets the session cookie.
func (h *HTTPHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if errParam := query.Get("error"); errParam != "" { // user pressed Cancel
		h.clearCookie(w, StateCookie)
		h.redirectHome(w, r, errParam)
		return
	}

	// Constant-time compare so the check cannot be probed byte by byte.
	expected := cookieValue(r, StateCookie)
	h.clearCookie(w, StateCookie)
	if expected == "" || subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(expected)) != 1 {
		h.redirectHome(w, r, "invalid_state")
		return
	}

	session, err := h.uc.CompleteSignIn(r.Context(), query.Get("code"))
	if err != nil {
		log.Printf("account: sign-in failed: %v", err)
		h.redirectHome(w, r, "login_failed")
		return
	}

	h.setCookie(w, SessionCookie, session.Token, h.opts.SessionMaxAge)
	h.redirectHome(w, r, "")
}

// Me returns the signed-in user, or 401 when there is no valid session.
func (h *HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.uc.Authenticate(r.Context(), bearerOrCookie(r))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "not authenticated"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// SignOut clears the session cookie. The JWT is stateless, so there is nothing
// to revoke server-side; it simply stops being presented.
func (h *HTTPHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	h.clearCookie(w, SessionCookie)
	writeJSON(w, http.StatusOK, map[string]any{"signed_out": true})
}

func (h *HTTPHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user, err := h.uc.Authenticate(r.Context(), bearerOrCookie(r))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "not authenticated"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		p, err := h.uc.ViewProfile(r.Context(), user.UserID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, p)

	case http.MethodPut, http.MethodPost:
		var req struct {
			DisplayName string `json:"display_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DisplayName) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "display_name is required"})
			return
		}
		p, err := h.uc.UpdateProfile(r.Context(), user.UserID, req.DisplayName)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, p)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (h *HTTPHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	user, err := h.uc.Authenticate(r.Context(), bearerOrCookie(r))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "not authenticated"})
		return
	}
	stats, err := h.uc.ViewStatistics(r.Context(), user.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
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

func (h *HTTPHandler) redirectHome(w http.ResponseWriter, r *http.Request, errCode string) {
	target := strings.TrimSuffix(h.opts.FrontendURL, "/") + "/"
	if errCode != "" {
		target += "?auth_error=" + errCode
	}
	http.Redirect(w, r, target, http.StatusFound)
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
