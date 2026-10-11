package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
)

// The proxies map /api/<service>/<path> onto the service's own prefix. A
// path that merely starts with the service name ("rewards") must survive.
func TestReverseProxy_MapsPathsBySegment(t *testing.T) {
	var got string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
	}))
	defer backend.Close()

	cfg := &config.Config{
		AccountServiceURL:     backend.URL,
		LeaderboardServiceURL: backend.URL,
		RewardServiceURL:      backend.URL,
	}
	mux := http.NewServeMux()
	handler.New(cfg, nil, nil).RegisterRoutes(mux)

	cases := map[string]string{
		"/api/reward/rewards":                 "/api/v1/reward/rewards",
		"/api/reward/all-rewards":             "/api/v1/reward/all-rewards",
		"/api/reward/v1/reward/rewards":       "/api/v1/reward/rewards",
		"/api/leaderboard":                    "/api/v1/leaderboard",
		"/api/leaderboard/leaderboard/weekly": "/api/v1/leaderboard/weekly",
		"/api/account/profile":                "/api/v1/account/profile",
	}
	for in, want := range cases {
		got = ""
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, in, nil))
		if got != want {
			t.Errorf("%s -> %q, want %q", in, got, want)
		}
	}
}
