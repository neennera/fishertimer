package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/account/internal/domain"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

// stubUsecase answers every call from fields, so each test sets only what it needs.
type stubUsecase struct {
	signIn     *domain.SignInResult
	signInErr  error
	signUp     *domain.Session
	signUpErr  error
	authUser   *domain.UserAccount
	authErr    error
	profileErr error
	pending    *domain.SignUpTicket
	pendingErr error
}

var _ usecase.Usecase = (*stubUsecase)(nil)

func (s *stubUsecase) SignInURL(state string) string {
	return "https://accounts.example/auth?state=" + state
}
func (s *stubUsecase) CompleteSignIn(context.Context, string) (*domain.SignInResult, error) {
	return s.signIn, s.signInErr
}
func (s *stubUsecase) PendingSignUp(string) (*domain.SignUpTicket, error) {
	return s.pending, s.pendingErr
}
func (s *stubUsecase) CompleteSignUp(context.Context, string, string) (*domain.Session, error) {
	return s.signUp, s.signUpErr
}
func (s *stubUsecase) Authenticate(context.Context, string) (*domain.UserAccount, error) {
	return s.authUser, s.authErr
}
func (s *stubUsecase) UpdateProfile(context.Context, string, string) (*domain.Session, error) {
	return s.signUp, s.signUpErr
}
func (s *stubUsecase) GetProfile(context.Context, string) (*domain.UserAccount, error) {
	return s.authUser, s.profileErr
}
func (s *stubUsecase) GetTimerStatistics(context.Context, string) (*domain.TimerStatistics, error) {
	return nil, s.profileErr
}
func (s *stubUsecase) GetRewards(context.Context, string) (*domain.RewardsSummary, error) {
	return nil, s.profileErr
}

func newServer(uc *stubUsecase) http.Handler {
	h := handler.New(uc, handler.Options{
		FrontendURL:     "http://web.test",
		OAuthConfigured: true,
		SessionMaxAge:   time.Hour,
	})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func do(srv http.Handler, method, target, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestWrongMethodIsRejected(t *testing.T) {
	srv := newServer(&stubUsecase{})
	cases := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/account/me"},
		{http.MethodDelete, "/api/v1/account/me"},
		{http.MethodPost, "/api/v1/account/google/login"},
		{http.MethodPost, "/api/v1/account/google/callback"},
		{http.MethodGet, "/api/v1/account/signout"},
		{http.MethodDelete, "/api/v1/account/signout"},
		{http.MethodGet, "/api/v1/account/signup"},
		{http.MethodPost, "/api/v1/account/update-profile"},
	}
	for _, c := range cases {
		if rec := do(srv, c.method, c.path, ""); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: expected 405, got %d", c.method, c.path, rec.Code)
		}
	}
}

func TestMe(t *testing.T) {
	user := &domain.UserAccount{UserID: "u", Email: "a@b.c", DisplayName: "A"}
	if rec := do(newServer(&stubUsecase{authUser: user}), http.MethodGet, "/api/v1/account/me", ""); rec.Code != http.StatusOK {
		t.Fatalf("signed in: expected 200, got %d", rec.Code)
	}
	if rec := do(newServer(&stubUsecase{authErr: domain.ErrUnauthorized}), http.MethodGet, "/api/v1/account/me", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("signed out: expected 401, got %d", rec.Code)
	}
}

func TestSignOutClearsCookies(t *testing.T) {
	rec := do(newServer(&stubUsecase{}), http.MethodPost, "/api/v1/account/signout", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	cleared := map[string]bool{}
	for _, c := range rec.Result().Cookies() {
		if c.MaxAge < 0 {
			cleared[c.Name] = true
		}
	}
	if !cleared[handler.SessionCookie] || !cleared[handler.SignUpCookie] {
		t.Fatalf("expected both cookies cleared, got %v", cleared)
	}
}

func TestSignUpStatusCodes(t *testing.T) {
	session := &domain.Session{Token: "tok", User: &domain.UserAccount{UserID: "u", DisplayName: "Fish"}}
	cases := []struct {
		name string
		uc   *stubUsecase
		body string
		want int
	}{
		{"created", &stubUsecase{signUp: session}, `{"display_name":"Fish"}`, http.StatusCreated},
		{"bad json", &stubUsecase{}, `nope`, http.StatusBadRequest},
		{"invalid name", &stubUsecase{signUpErr: domain.ErrInvalid}, `{"display_name":""}`, http.StatusBadRequest},
		{"ticket expired", &stubUsecase{signUpErr: domain.ErrUnauthorized}, `{"display_name":"x"}`, http.StatusUnauthorized},
		{"other failure", &stubUsecase{signUpErr: context.DeadlineExceeded}, `{"display_name":"x"}`, http.StatusInternalServerError},
	}
	for _, c := range cases {
		rec := do(newServer(c.uc), http.MethodPost, "/api/v1/account/signup", c.body)
		if rec.Code != c.want {
			t.Errorf("%s: expected %d, got %d (%s)", c.name, c.want, rec.Code, rec.Body.String())
		}
	}

	rec := do(newServer(&stubUsecase{signUp: session}), http.MethodPost, "/api/v1/account/signup", `{"display_name":"Fish"}`)
	var set bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == handler.SessionCookie && c.Value == "tok" && c.HttpOnly {
			set = true
		}
	}
	if !set {
		t.Fatal("expected an HttpOnly ft_session cookie on success")
	}
}

func TestUpdateProfileStatusCodes(t *testing.T) {
	session := &domain.Session{Token: "tok", User: &domain.UserAccount{UserID: "u", DisplayName: "New"}}
	cases := []struct {
		name string
		uc   *stubUsecase
		want int
	}{
		{"ok", &stubUsecase{signUp: session}, http.StatusOK},
		{"invalid", &stubUsecase{signUpErr: domain.ErrInvalid}, http.StatusBadRequest},
		{"signed out", &stubUsecase{signUpErr: domain.ErrUnauthorized}, http.StatusUnauthorized},
		{"user gone", &stubUsecase{signUpErr: domain.ErrNotFound}, http.StatusUnauthorized},
		{"other", &stubUsecase{signUpErr: context.DeadlineExceeded}, http.StatusInternalServerError},
	}
	for _, c := range cases {
		rec := do(newServer(c.uc), http.MethodPatch, "/api/v1/account/update-profile", `{"display_name":"New"}`)
		if rec.Code != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, rec.Code)
		}
	}
}

func TestProfileLookupStatusCodes(t *testing.T) {
	user := &domain.UserAccount{UserID: "u"}
	cases := []struct {
		name, path string
		uc         *stubUsecase
		want       int
	}{
		{"ok", "/api/v1/account/profile?id=x", &stubUsecase{authUser: user}, http.StatusOK},
		{"missing id", "/api/v1/account/profile", &stubUsecase{profileErr: domain.ErrInvalid}, http.StatusBadRequest},
		{"unknown", "/api/v1/account/profile?id=x", &stubUsecase{profileErr: domain.ErrNotFound}, http.StatusNotFound},
		{"stats unknown", "/api/v1/account/statistics?id=x", &stubUsecase{profileErr: domain.ErrNotFound}, http.StatusNotFound},
		{"stats upstream", "/api/v1/account/statistics?id=x", &stubUsecase{profileErr: context.DeadlineExceeded}, http.StatusBadGateway},
		{"rewards upstream", "/api/v1/account/rewards?id=x", &stubUsecase{profileErr: context.DeadlineExceeded}, http.StatusBadGateway},
	}
	for _, c := range cases {
		if rec := do(newServer(c.uc), http.MethodGet, c.path, ""); rec.Code != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, rec.Code)
		}
	}
}

func TestGoogleLogin_OpenRedirectIsNeutralised(t *testing.T) {
	rec := do(newServer(&stubUsecase{}), http.MethodGet, "/api/v1/account/google/login?next=//evil.example&signup=https://evil.example", "")
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case handler.NextCookie:
			if c.Value != handler.DefaultNextPath {
				t.Errorf("next cookie = %q, want the default", c.Value)
			}
		case handler.SignUpPathCookie:
			if c.Value != handler.DefaultSignUpPath {
				t.Errorf("signup cookie = %q, want the default", c.Value)
			}
		}
	}
}

func location(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
	return rec.Header().Get("Location")
}

func TestGoogleCallback_ErrorsReturnToRememberedPage(t *testing.T) {
	srv := newServer(&stubUsecase{signInErr: context.DeadlineExceeded})
	next := &http.Cookie{Name: handler.NextCookie, Value: "/signin"}

	got := location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?error=access_denied", "", next))
	if got != "http://web.test/signin?auth_error=access_denied" {
		t.Errorf("cancel: got %s", got)
	}

	got = location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?code=c&state=wrong", "", next,
		&http.Cookie{Name: handler.StateCookie, Value: "right"}))
	if got != "http://web.test/signin?auth_error=invalid_state" {
		t.Errorf("bad state: got %s", got)
	}

	got = location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?code=c&state=s", "", next,
		&http.Cookie{Name: handler.StateCookie, Value: "s"}))
	if got != "http://web.test/signin?auth_error=login_failed" {
		t.Errorf("exchange failed: got %s", got)
	}
}

func TestGoogleCallback_LostCookiesStillLandOnSignIn(t *testing.T) {
	// The cookies live 10 minutes; after a slow consent screen they are gone.
	srv := newServer(&stubUsecase{})
	got := location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?code=c&state=s", ""))
	if got != "http://web.test"+handler.DefaultErrorPath+"?auth_error=invalid_state" {
		t.Errorf("expected the sign-in page, got %s", got)
	}
}

func TestGoogleCallback_SuccessPaths(t *testing.T) {
	state := &http.Cookie{Name: handler.StateCookie, Value: "s"}
	next := &http.Cookie{Name: handler.NextCookie, Value: "/rooms"}
	signup := &http.Cookie{Name: handler.SignUpPathCookie, Value: "/welcome"}

	existing := &domain.Session{Token: "tok", User: &domain.UserAccount{Email: "a@b.c"}}
	srv := newServer(&stubUsecase{signIn: &domain.SignInResult{Session: existing}})
	got := location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?code=c&state=s", "", state, next, signup))
	if got != "http://web.test/rooms" {
		t.Errorf("known e-mail: got %s", got)
	}

	ticket := &domain.SignUpTicket{Email: "new@b.c", ExpiresAt: time.Now().Add(time.Minute)}
	srv = newServer(&stubUsecase{signIn: &domain.SignInResult{SignUpToken: "t", SignUp: ticket}})
	got = location(t, do(srv, http.MethodGet, "/api/v1/account/google/callback?code=c&state=s", "", state, next, signup))
	if got != "http://web.test/welcome" {
		t.Errorf("new e-mail: got %s", got)
	}
}

// Anyone can call /profile, so it must not carry the e-mail or the role.
func TestProfile_IsThePublicViewOnly(t *testing.T) {
	user := &domain.UserAccount{
		UserID: "u1", Email: "private@example.com", DisplayName: "Fish", AvatarURL: "https://img", Role: domain.RoleAdmin,
	}
	rec := do(newServer(&stubUsecase{authUser: user}), http.MethodGet, "/api/v1/account/profile?id=u1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, secret := range []string{"private@example.com", "email", "role", "ADMIN"} {
		if strings.Contains(body, secret) {
			t.Fatalf("public profile leaks %q: %s", secret, body)
		}
	}
	for _, want := range []string{`"user_id":"u1"`, `"display_name":"Fish"`, `"avatar_url":"https://img"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("public profile is missing %s: %s", want, body)
		}
	}
	// /me is the owner's own view and keeps the e-mail.
	me := do(newServer(&stubUsecase{authUser: user}), http.MethodGet, "/api/v1/account/me", "")
	if !strings.Contains(me.Body.String(), "private@example.com") {
		t.Fatalf("/me must still return the owner's e-mail: %s", me.Body.String())
	}
}

func TestPendingSignUp_OffersTheGoogleName(t *testing.T) {
	ticket := &domain.SignUpTicket{Email: "new@example.com", Name: "Mali Chai", Picture: "https://img", ExpiresAt: time.Now().Add(time.Minute)}
	rec := do(newServer(&stubUsecase{pending: ticket}), http.MethodGet, "/api/v1/account/signup/pending", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body)
	}
	for _, want := range []string{`"email":"new@example.com"`, `"suggested_name":"Mali Chai"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("missing %s in %s", want, rec.Body)
		}
	}
}

func TestPendingSignUp_ExpiredAndWrongMethod(t *testing.T) {
	if rec := do(newServer(&stubUsecase{pendingErr: domain.ErrUnauthorized}), http.MethodGet, "/api/v1/account/signup/pending", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired ticket: expected 401, got %d", rec.Code)
	}
	if rec := do(newServer(&stubUsecase{}), http.MethodPost, "/api/v1/account/signup/pending", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST: expected 405, got %d", rec.Code)
	}
}

func TestDefaultSignUpPathIsTheWelcomePage(t *testing.T) {
	if handler.DefaultSignUpPath != "/welcome" {
		t.Fatalf("the sign-up default must be the page that exists, got %q", handler.DefaultSignUpPath)
	}
}
