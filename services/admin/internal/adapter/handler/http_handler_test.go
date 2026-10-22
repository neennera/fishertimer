package handler_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/middleware"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

const (
	secret  = "test-secret"
	issuer  = "fishertimer-account"
	adminID = "57f2060b-edf9-499e-a430-9d6565591b25"
	roomID  = "f2000000-0000-4000-8000-000000000001"
	userID  = "e1000000-0000-4000-8000-000000000002"
)

func token(t *testing.T, role string, exp time.Time) string {
	t.Helper()
	enc := base64.RawURLEncoding
	head := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, _ := json.Marshal(map[string]any{
		"sub": adminID, "role": role, "use": "session", "iss": issuer, "exp": exp.Unix(),
	})
	unsigned := head + "." + enc.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc.EncodeToString(mac.Sum(nil))
}

type liveSessions struct{}

func (liveSessions) ListActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	return []domain.SessionOverview{{SessionID: roomID, IsActive: true}}, nil
}

func (liveSessions) GetActiveSession(ctx context.Context, id string) (*domain.SessionOverview, error) {
	if id != roomID {
		return nil, domain.ErrNotFound
	}
	return &domain.SessionOverview{SessionID: roomID, IsActive: true}, nil
}

func (liveSessions) ListParticipants(ctx context.Context, id string) ([]domain.SessionParticipantOverview, error) {
	return nil, nil
}

type recordingClient struct{ kickedUser, closedRoom string }

func (c *recordingClient) KickParticipant(ctx context.Context, sessionID, userID string) (bool, error) {
	c.kickedUser = sessionID + "/" + userID
	return false, nil
}

func (c *recordingClient) CloseSession(ctx context.Context, sessionID string) error {
	c.closedRoom = sessionID
	return nil
}

func newServer() (*http.ServeMux, *recordingClient, *repository.InMemoryRepository) {
	logs := repository.NewInMemory()
	client := &recordingClient{}
	h := handler.New(usecase.New(liveSessions{}, logs, client), middleware.NewVerifier(secret, issuer))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux, client, logs
}

func do(mux http.Handler, method, path, tok, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if tok != "" {
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: tok})
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAdminRoutes_RequireAdminRole(t *testing.T) {
	mux, _, _ := newServer()
	cases := []struct {
		name string
		tok  string
		want int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"expired", token(t, "ADMIN", time.Now().Add(-time.Minute)), http.StatusUnauthorized},
		{"tampered", token(t, "ADMIN", time.Now().Add(time.Hour)) + "x", http.StatusUnauthorized},
		{"customer", token(t, "CUSTOMER", time.Now().Add(time.Hour)), http.StatusForbidden},
		{"admin", token(t, "ADMIN", time.Now().Add(time.Hour)), http.StatusOK},
	}
	for _, c := range cases {
		if rec := do(mux, http.MethodGet, "/api/v1/admin/sessions", c.tok, ""); rec.Code != c.want {
			t.Errorf("%s: expected %d, got %d (%s)", c.name, c.want, rec.Code, rec.Body)
		}
	}
	if rec := do(mux, http.MethodGet, "/health", "", ""); rec.Code != http.StatusOK {
		t.Errorf("health must stay public, got %d", rec.Code)
	}
}

func TestKickRoute_UsesPathAndAuditsAsAdmin(t *testing.T) {
	mux, client, logs := newServer()
	tok := token(t, "ADMIN", time.Now().Add(time.Hour))

	rec := do(mux, http.MethodPost, "/api/v1/admin/sessions/"+roomID+"/participants/"+userID+"/kick", tok, `{"reason":"spam"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body)
	}
	if client.kickedUser != roomID+"/"+userID {
		t.Fatalf("unexpected kick target %q", client.kickedUser)
	}
	stored, _ := logs.List(context.Background(), domain.ActionKickUser, 10)
	if len(stored) != 1 || stored[0].AdminID != adminID || stored[0].TargetID != userID {
		t.Fatalf("unexpected audit log %+v", stored)
	}
}

func TestEndRoute_EmptyBodyAndEndedRoom(t *testing.T) {
	mux, client, _ := newServer()
	tok := token(t, "ADMIN", time.Now().Add(time.Hour))

	if rec := do(mux, http.MethodPost, "/api/v1/admin/sessions/"+roomID+"/end", tok, ""); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body)
	}
	if client.closedRoom != roomID {
		t.Fatalf("unexpected closed room %q", client.closedRoom)
	}
	other := "f2000000-0000-4000-8000-000000000009"
	if rec := do(mux, http.MethodPost, "/api/v1/admin/sessions/"+other+"/end", tok, ""); rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for a room that is not live, got %d", rec.Code)
	}
}

// The page shows the error text as is, so it must be a sentence for the admin:
// no "admin: invalid input:" style prefixes from the sentinel errors.
func TestErrorMessagesAreSentencesWithoutInternalPrefixes(t *testing.T) {
	mux, _, _ := newServer()
	tok := token(t, "ADMIN", time.Now().Add(time.Hour))
	other := "f2000000-0000-4000-8000-000000000009"

	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"bad ids", http.MethodPost, "/api/v1/admin/sessions/abc/participants/def/kick", "", http.StatusBadRequest},
		{"bad room id", http.MethodPost, "/api/v1/admin/sessions/abc/end", "", http.StatusBadRequest},
		{"room ended", http.MethodPost, "/api/v1/admin/sessions/" + other + "/end", "", http.StatusConflict},
		{"long reason", http.MethodPost, "/api/v1/admin/sessions/" + roomID + "/end", `{"reason":"` + strings.Repeat("x", 501) + `"}`, http.StatusBadRequest},
		{"unknown action", http.MethodGet, "/api/v1/admin/logs?action=BAN", "", http.StatusBadRequest},
	}
	for _, c := range cases {
		rec := do(mux, c.method, c.path, tok, c.body)
		if rec.Code != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, rec.Code)
			continue
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
			t.Errorf("%s: expected an error body, got %q", c.name, rec.Body.String())
			continue
		}
		if strings.Contains(body.Error, "admin:") || strings.Contains(body.Error, "study-session") {
			t.Errorf("%s: internal prefix leaked: %q", c.name, body.Error)
		}
	}
}
