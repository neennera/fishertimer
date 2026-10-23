package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbe(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()

	if err := probe(ok.URL); err != nil {
		t.Fatalf("2xx should be healthy, got %v", err)
	}
	if err := probe(down.URL); err == nil {
		t.Fatal("503 should be unhealthy")
	}
	closed := httptest.NewServer(http.NotFoundHandler())
	url := closed.URL
	closed.Close()
	if err := probe(url); err == nil {
		t.Fatal("a closed port should be unhealthy")
	}
}
