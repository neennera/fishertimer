package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

type GatewayHandler struct {
	cfg          *config.Config
	accountProxy *httputil.ReverseProxy
	timer        *TimerHandler
	boardProxy   *httputil.ReverseProxy
	session      *SessionHandler
	rewardProxy  *httputil.ReverseProxy
}

// New builds the gateway. Most services are reverse-proxied over HTTP; Study
// Timer and Study Session are reached over gRPC through their clients.
func New(cfg *config.Config, timerClient timerv1.StudyTimerServiceClient, sessionClient sessionv1.StudySessionServiceClient) *GatewayHandler {
	return &GatewayHandler{
		cfg:          cfg,
		accountProxy: createReverseProxy(cfg.AccountServiceURL, "/api/v1/account"),
		timer:        NewTimerHandler(timerClient, cfg.Env != "production"),
		boardProxy:   createReverseProxy(cfg.LeaderboardServiceURL, "/api/v1/leaderboard"),
		session:      NewSessionHandler(sessionClient, cfg.Env != "production"),
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
		path := strings.TrimPrefix(req.URL.Path, "/api/")
		parts := strings.SplitN(path, "/", 2)

		subPath := ""
		if len(parts) == 2 {
			subPath = strings.TrimPrefix(parts[1], "/")
		}

		// Normalize: strip accidental duplicate service prefixes like "v1/leaderboard" or "leaderboard"
		prefixNoSlash := strings.TrimPrefix(prefix, "/")
		v1Prefix := strings.TrimPrefix(prefixNoSlash, "api/")
		serviceName := parts[0]

		subPath = trimSegments(subPath, prefixNoSlash)
		subPath = trimSegments(subPath, v1Prefix)
		subPath = trimSegments(subPath, serviceName)

		if subPath != "" {
			req.URL.Path = prefix + "/" + subPath
		} else {
			req.URL.Path = prefix
		}
	}
	return proxy
}

// trimSegments removes prefix from path only when it is made of whole path
// segments: "reward/x" loses "reward", but "rewards" keeps its name (a plain
// string trim turned /api/reward/rewards into /api/v1/reward/s).
func trimSegments(path, prefix string) string {
	switch {
	case prefix == "":
		return path
	case path == prefix:
		return ""
	case strings.HasPrefix(path, prefix+"/"):
		return strings.TrimPrefix(path[len(prefix)+1:], "/")
	default:
		return path
	}
}

func (h *GatewayHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.withCORS(h.Health))
	mux.HandleFunc("/api/v1/gateway/status", h.withCORS(h.Status))
	mux.HandleFunc("/api/v1/gateway/whoami", h.withCORS(h.Whoami))

	// Microservices proxy routes for Client
	mux.HandleFunc("/api/account", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/account/", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/auth", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/auth/", h.withCORS(h.handleAccount))
	mux.HandleFunc("/api/timer", h.withCORS(h.handleTimer))
	mux.HandleFunc("/api/timer/", h.withCORS(h.handleTimer))
	mux.HandleFunc("/api/leaderboard", h.withCORS(h.handleLeaderboard))
	mux.HandleFunc("/api/leaderboard/", h.withCORS(h.handleLeaderboard))
	mux.HandleFunc("/api/session", h.withCORS(h.handleSession))
	mux.HandleFunc("/api/session/", h.withCORS(h.handleSession))
	mux.HandleFunc("/api/reward", h.withCORS(h.handleReward))
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

// Whoami echoes the identity headers middleware.Verifier.Identity attached to
// this request, purely so the header-forwarding contract can be exercised
// end-to-end from the browser without waiting on a downstream service to
// start reading them.
func (h *GatewayHandler) Whoami(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(middleware.HeaderUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"authenticated":  userID != "",
		"x_user_id":      userID,
		"x_user_role":    r.Header.Get(middleware.HeaderUserRole),
		"x_display_name": r.Header.Get(middleware.HeaderDisplayName),
	})
}

func (h *GatewayHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	h.accountProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleTimer(w http.ResponseWriter, r *http.Request) {
	h.timer.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.boardProxy.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleSession(w http.ResponseWriter, r *http.Request) {
	h.session.ServeHTTP(w, r)
}

func (h *GatewayHandler) handleReward(w http.ResponseWriter, r *http.Request) {
	h.rewardProxy.ServeHTTP(w, r)
}
