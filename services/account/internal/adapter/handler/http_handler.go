package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/account/status", h.Status)
	mux.HandleFunc("/api/v1/account/signin", h.SignIn)
	mux.HandleFunc("/api/v1/account/signup", h.SignUp)
	mux.HandleFunc("/api/v1/account/signout", h.SignOut)
	mux.HandleFunc("/api/v1/account/profile", h.Profile)
	mux.HandleFunc("/api/v1/account/statistics", h.Statistics)
}

type authRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type profileUpdateRequest struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

func (h *HTTPHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	u, err := h.uc.SignIn(r.Context(), req.Email, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *HTTPHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	u, err := h.uc.SignUp(r.Context(), req.Email, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *HTTPHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "signed_out"})
}

func (h *HTTPHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if r.Method == http.MethodGet {
		p, err := h.uc.ViewProfile(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	}

	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		var req profileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		p, err := h.uc.UpdateProfile(r.Context(), req.UserID, req.DisplayName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (h *HTTPHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	stats, err := h.uc.ViewStatistics(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "account",
		"status":    "healthy",
		"port":      8082,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "account",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
