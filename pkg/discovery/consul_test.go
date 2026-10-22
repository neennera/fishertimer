package discovery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConsulResolveReturnsPassingInstances(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		io.WriteString(w, `[
		  {"Node":{"Address":"10.0.0.9"},"Service":{"ID":"a","Service":"account","Address":"10.0.0.5","Port":8082}},
		  {"Node":{"Address":"10.0.0.9"},"Service":{"ID":"b","Service":"account","Address":"","Port":8082}}
		]`)
	}))
	defer srv.Close()

	got, err := NewConsul(srv.URL).Resolve(context.Background(), "account")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/health/service/account" || gotQuery != "passing=true" {
		t.Fatalf("asked %s?%s", gotPath, gotQuery)
	}
	if len(got) != 2 || got[0].HostPort() != "10.0.0.5:8082" {
		t.Fatalf("got %+v", got)
	}
	// An empty service address means "the node's address".
	if got[1].HostPort() != "10.0.0.9:8082" {
		t.Fatalf("node address not used: %+v", got[1])
	}
}

func TestConsulResolveErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	if _, err := NewConsul(srv.URL).Resolve(context.Background(), "x"); err == nil {
		t.Fatal("a 500 must be an error")
	}
	srv.Close()
	if _, err := NewConsul(srv.URL).Resolve(context.Background(), "x"); err == nil {
		t.Fatal("an unreachable agent must be an error")
	}
}

func TestConsulRegisterAndDeregister(t *testing.T) {
	var body map[string]any
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/register") {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
		}
	}))
	defer srv.Close()

	c := NewConsul(strings.TrimPrefix(srv.URL, "http://")) // host:port form
	reg := HTTPService("account", "10.0.0.5", 8082)
	if err := c.Register(context.Background(), reg); err != nil {
		t.Fatal(err)
	}
	if err := c.Deregister(context.Background(), reg.ID); err != nil {
		t.Fatal(err)
	}

	if body["Name"] != "account" || body["Address"] != "10.0.0.5" || body["Port"].(float64) != 8082 {
		t.Fatalf("body %v", body)
	}
	chk := body["Check"].(map[string]any)
	if chk["HTTP"] != "http://10.0.0.5:8082/health" || chk["Interval"] != "5s" || chk["DeregisterCriticalServiceAfter"] != "1m0s" {
		t.Fatalf("check %v", chk)
	}
	want := []string{"PUT /v1/agent/service/register", "PUT /v1/agent/service/deregister/account-10.0.0.5-8082"}
	if len(paths) != 2 || paths[0] != want[0] || paths[1] != want[1] {
		t.Fatalf("calls %v", paths)
	}
}

func TestGRPCServiceUsesTCPCheckUnderGRPCName(t *testing.T) {
	reg := GRPCService("study-timer", "10.0.0.7", 50051)
	if reg.Name != "study-timer-grpc" || reg.Check.TCP != "10.0.0.7:50051" || reg.Check.HTTP != "" {
		t.Fatalf("%+v", reg)
	}
	if got := reg.Check.withDefaults(); got.Interval != 5*time.Second || got.Timeout != 2*time.Second {
		t.Fatalf("defaults %+v", got)
	}
}
