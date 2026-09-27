package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Gateway Handler
	h := handler.New(cfg)

	// 2. Setup Router & Server
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Verifies the account service's session JWT (cookie or bearer header)
	// and forwards trusted X-User-* headers downstream. It never rejects a
	// request itself - enforcing that a route requires a session is left to
	// each downstream service.
	verifier := middleware.NewVerifier(cfg.JWTSecret, cfg.JWTIssuer)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      verifier.Identity(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("api-gateway service listening on port %d [%s]", cfg.Port, cfg.Env)
		log.Printf("  -> Account:     %s", cfg.AccountServiceURL)
		log.Printf("  -> Timer:       %s", cfg.TimerServiceURL)
		log.Printf("  -> Leaderboard: %s", cfg.LeaderboardServiceURL)
		log.Printf("  -> Session:     %s", cfg.SessionServiceURL)
		log.Printf("  -> Reward:      %s", cfg.RewardServiceURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down api-gateway service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("api-gateway exited cleanly")
}
