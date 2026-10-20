package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neennera/fishertimer/services/leaderboard/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
)

type stubUsecase struct {
	result *domain.CachedRanking
	err    error
	period string
}

func (s *stubUsecase) ViewLeaderboard(_ context.Context, period string) (*domain.CachedRanking, error) {
	s.period = period
	return s.result, s.err
}

func get(uc *stubUsecase, method, target string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	handler.New(uc).RegisterRoutes(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func errorBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("error must be JSON, got Content-Type %q (%s)", ct, rec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
		t.Fatalf("expected {\"error\": ...}, got %q (%v)", rec.Body.String(), err)
	}
	return body.Error
}

func TestViewLeaderboard_OK(t *testing.T) {
	uc := &stubUsecase{result: &domain.CachedRanking{
		Period:   "weekly",
		Rankings: []domain.RankEntry{{UserID: "u", DisplayName: "A", Rank: 1, Score: 10}},
	}}
	rec := get(uc, http.MethodGet, "/api/v1/leaderboard?period=weekly")
	if rec.Code != http.StatusOK || uc.period != "weekly" {
		t.Fatalf("expected 200 for weekly, got %d (period %q)", rec.Code, uc.period)
	}
}

func TestViewLeaderboard_MissingPeriodDefaultsToAllTime(t *testing.T) {
	uc := &stubUsecase{result: &domain.CachedRanking{}}
	get(uc, http.MethodGet, "/api/v1/leaderboard")
	if uc.period != "all-time" {
		t.Fatalf("expected all-time, got %q", uc.period)
	}
}

func TestViewLeaderboard_EmptyIsAnEmptyArrayNotNull(t *testing.T) {
	// E-1: no rewards yet is a normal answer, and the client iterates the array.
	uc := &stubUsecase{result: &domain.CachedRanking{Period: "monthly"}}
	for _, path := range []string{"/api/v1/leaderboard?period=monthly", "/api/v1/leaderboard/rankings?period=monthly"} {
		rec := get(uc, http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "null") {
			t.Fatalf("%s: rankings must be [] not null: %s", path, rec.Body.String())
		}
	}
}

func TestViewLeaderboard_RewardServiceDownIs502WithSafeJSON(t *testing.T) {
	secret := "dial tcp 10.0.0.5:8085: connect: connection refused"
	uc := &stubUsecase{err: fmt.Errorf("%w: %s", domain.ErrUpstream, secret)}
	for _, path := range []string{"/api/v1/leaderboard?period=weekly", "/api/v1/leaderboard/rankings"} {
		rec := get(uc, http.MethodGet, path)
		if rec.Code != http.StatusBadGateway {
			t.Fatalf("%s: expected 502, got %d", path, rec.Code)
		}
		if msg := errorBody(t, rec); strings.Contains(msg, "10.0.0.5") || strings.Contains(msg, "dial") {
			t.Fatalf("%s: internal error text leaked to the client: %q", path, msg)
		}
	}
}

func TestViewLeaderboard_OtherFailureIs500WithSafeJSON(t *testing.T) {
	rec := get(&stubUsecase{err: errors.New("boom: secret detail")}, http.MethodGet, "/api/v1/leaderboard")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if msg := errorBody(t, rec); strings.Contains(msg, "secret") {
		t.Fatalf("internal error text leaked: %q", msg)
	}
}

func TestWrongMethodIsJSON405(t *testing.T) {
	for _, path := range []string{"/api/v1/leaderboard", "/api/v1/leaderboard/rankings"} {
		rec := get(&stubUsecase{}, http.MethodPost, path)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s: expected 405, got %d", path, rec.Code)
		}
		errorBody(t, rec)
	}
}

func TestOptionsPreflight(t *testing.T) {
	rec := get(&stubUsecase{}, http.MethodOptions, "/api/v1/leaderboard")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
