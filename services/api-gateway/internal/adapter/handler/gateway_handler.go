package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/neennera/fishertimer/services/api-gateway/config"
)

type GatewayHandler struct {
	cfg          *config.Config
	accountProxy *httputil.ReverseProxy
	timerProxy   *httputil.ReverseProxy
	boardProxy   *httputil.ReverseProxy
	sessionProxy *httputil.ReverseProxy
	rewardProxy  *httputil.ReverseProxy
}

func New(cfg *config.Config) *GatewayHandler {
	return &GatewayHandler{
		cfg:          cfg,
		accountProxy: createReverseProxy(cfg.AccountServiceURL, "/api/v1/account"),
		timerProxy:   createReverseProxy(cfg.TimerServiceURL, "/api/v1/study-timer"),
		boardProxy:   createReverseProxy(cfg.LeaderboardServiceURL, "/api/v1/leaderboard"),
		sessionProxy: createReverseProxy(cfg.SessionServiceURL, "/api/v1/study-session"),
		rewardProxy:  createReverseProxy(cfg.RewardServiceURL, "/api/v1/reward"),
	}
}

func createReverseProxy(targetURL, prefix string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host

		// Map /api/<service>/... to <prefix>/...
		path := req.URL.Path
		parts := strings.SplitN(strings.TrimPrefix(path, "/api/"), "/", 2)
		if len(parts) == 2 {
			req.URL.Path = prefix + "/" + parts[1]
		} else {
			req.URL.Path = prefix
		}
	}
	return proxy
}

func (h *GatewayHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.withCORS(h.Health))
	mux.HandleFunc("/api/v1/gateway/status", h.withCORS(h.Status))

	// Microservices proxy routes for Client
	mux.HandleFunc("/api/account/", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/auth/", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/timer/", h.withCORS(h.handleTimer))
	mux.HandleFunc("/api/leaderboard/", h.withCORS(h.handleLeaderboard))
	mux.HandleFunc("/api/session/", h.withCORS(h.handleSession))
	mux.HandleFunc("/api/reward/", h.withCORS(h.handleReward))
}

func (h *GatewayHandler) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (h *GatewayHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "api-gateway",
		"status":    "healthy",
		"port":      h.cfg.Port,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *GatewayHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "api-gateway",
		"layer":   "adapter.handler",
		"routes": []string{
			"/api/account/*",
			"/api/auth/*",
			"/api/timer/*",
			"/api/leaderboard/*",
			"/api/session/*",
			"/api/reward/*",
		},
		"ready": true,
	})
}

func (h *GatewayHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	h.accountProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleTimer(w http.ResponseWriter, r *http.Request) {
	h.timerProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.boardProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleSession(w http.ResponseWriter, r *http.Request) {
	h.sessionProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleReward(w http.ResponseWriter, r *http.Request) {
	h.rewardProxy.ServeHTTP(w, r)
}
