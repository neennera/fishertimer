package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/neennera/fishertimer/pkg/discovery"
	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
)

// registry is a discovery.Source that always answers with the same instances.
type registry map[string][]discovery.Instance

func (r registry) Resolve(_ context.Context, service string) ([]discovery.Instance, error) {
	return r[service], nil
}

func instanceOf(t *testing.T, srv *httptest.Server) discovery.Instance {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return discovery.Instance{Host: u.Hostname(), Port: port}
}

// With discovery on, a proxy sends the request to the instance the registry
// names, not to the address in the configuration (which here points nowhere).
func TestReverseProxy_FollowsDiscoveredInstances(t *testing.T) {
	var gotPath string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
	}))
	defer backend.Close()

	deadAddress := "http://127.0.0.1:1"
	cfg := &config.Config{
		AccountServiceURL:     deadAddress,
		LeaderboardServiceURL: deadAddress,
		RewardServiceURL:      deadAddress,
	}
	loc := discovery.NewLocator(
		registry{discovery.Reward: {instanceOf(t, backend)}},
		map[string]string{discovery.Reward: deadAddress},
		discovery.WithLogf(func(string, ...any) {}),
	)
	mux := http.NewServeMux()
	handler.New(cfg, nil, nil, handler.WithDiscovery(loc)).RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/reward/rewards", nil))
	if gotPath != "/api/v1/reward/rewards" {
		t.Fatalf("the discovered instance did not get the request (status %d, path %q)", rec.Code, gotPath)
	}
}

// A static Locator (no registry) changes nothing: requests go to the
// configured address.
func TestReverseProxy_StaticDiscoveryUsesConfiguredAddress(t *testing.T) {
	var gotPath string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
	}))
	defer backend.Close()

	cfg := &config.Config{RewardServiceURL: backend.URL}
	loc := discovery.NewLocator(nil, map[string]string{discovery.Reward: backend.URL})
	mux := http.NewServeMux()
	handler.New(cfg, nil, nil, handler.WithDiscovery(loc)).RegisterRoutes(mux)

	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/reward/rewards", nil))
	if gotPath != "/api/v1/reward/rewards" {
		t.Fatalf("path %q", gotPath)
	}
}
