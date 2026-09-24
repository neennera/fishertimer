package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
)

func TestGatewayHandler_Health(t *testing.T) {
	cfg := &config.Config{
		Port:                  8080,
		Env:                   "test",
		AccountServiceURL:     "http://localhost:8082",
		TimerServiceURL:       "http://localhost:8084",
		LeaderboardServiceURL: "http://localhost:8086",
		SessionServiceURL:     "http://localhost:8083",
		RewardServiceURL:      "http://localhost:8085",
	}

	h := handler.New(cfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestGatewayHandler_Status(t *testing.T) {
	cfg := &config.Config{
		Port:                  8080,
		Env:                   "test",
		AccountServiceURL:     "http://localhost:8082",
		TimerServiceURL:       "http://localhost:8084",
		LeaderboardServiceURL: "http://localhost:8086",
		SessionServiceURL:     "http://localhost:8083",
		RewardServiceURL:      "http://localhost:8085",
	}

	h := handler.New(cfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gateway/status", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
