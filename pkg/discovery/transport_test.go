package discovery

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func instanceOf(t *testing.T, srv *httptest.Server) Instance {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return Instance{Host: u.Hostname(), Port: port}
}

func TestTransportIsInertInStaticMode(t *testing.T) {
	l := NewLocator(nil, map[string]string{Reward: "http://reward:8085"}, quiet())
	if got := l.Transport(Reward, nil); got != http.DefaultTransport {
		t.Fatalf("static mode must leave the transport alone, got %T", got)
	}
	c := l.HTTPClient(Reward)
	if c.Timeout != DefaultTimeout || c.Transport != http.DefaultTransport {
		t.Fatalf("client %+v", c)
	}
	var nilLocator *Locator
	if nilLocator.Enabled() {
		t.Fatal("a nil locator is static")
	}
}

func TestTransportSendsRequestsToHealthyInstances(t *testing.T) {
	hit := map[string]int{}
	mk := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hit[name]++
			io.WriteString(w, name+" "+r.URL.RequestURI())
		}))
	}
	a, b := mk("a"), mk("b")
	defer a.Close()
	defer b.Close()

	src := &fakeSource{out: map[string][]Instance{Reward: {instanceOf(t, a), instanceOf(t, b)}}}
	l := NewLocator(src, map[string]string{Reward: "http://reward.invalid:8085"}, quiet(), WithTTL(time.Minute))
	client := l.HTTPClient(Reward)

	// The configured base URL points nowhere; discovery decides where it goes.
	var bodies []string
	for i := 0; i < 4; i++ {
		resp, err := client.Get("http://reward.invalid:8085/api/v1/reward/rewards?user_id=u1")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodies = append(bodies, string(body))
	}
	if hit["a"] != 2 || hit["b"] != 2 {
		t.Fatalf("calls should alternate between instances: %v", hit)
	}
	if bodies[0] != "a /api/v1/reward/rewards?user_id=u1" && bodies[0] != "b /api/v1/reward/rewards?user_id=u1" {
		t.Fatalf("path and query must survive: %q", bodies[0])
	}
}

func TestTransportDoesNotModifyTheCallersRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	src := &fakeSource{out: map[string][]Instance{Reward: {instanceOf(t, srv)}}}
	l := NewLocator(src, nil, quiet())

	req, _ := http.NewRequest(http.MethodGet, "http://reward.invalid:8085/x", nil)
	resp, err := l.Transport(Reward, nil).RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if req.URL.Host != "reward.invalid:8085" {
		t.Fatalf("request was modified: %s", req.URL.Host)
	}
}

func TestTransportUsesConfiguredAddressWhenRegistryIsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "static") }))
	defer srv.Close()
	src := &fakeSource{out: map[string][]Instance{}}
	l := NewLocator(src, map[string]string{Reward: srv.URL}, quiet())

	resp, err := l.HTTPClient(Reward).Get(srv.URL + "/ping")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "static" {
		t.Fatalf("got %q", body)
	}
}
