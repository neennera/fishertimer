package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/leaderboard/status", h.Status)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "leaderboard",
		"status":    "healthy",
		"port":      8086,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "leaderboard",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
