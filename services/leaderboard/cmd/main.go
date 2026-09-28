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

	"github.com/neennera/fishertimer/services/leaderboard/config"
	"github.com/neennera/fishertimer/services/leaderboard/internal/adapter/client"
	"github.com/neennera/fishertimer/services/leaderboard/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/leaderboard/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

func main() {
	cfg := config.Load()

	// ── 1. Repository (Redis if available, in-memory fallback) ───────────────
	var repo domain.Repository
	ctx := context.Background()

	redisRepo, err := repository.NewRedisFromURL(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("⚠️  Redis unavailable (%v) — falling back to in-memory cache", err)
		repo = repository.NewInMemory()
	} else {
		log.Printf("✅ Connected to Redis at %s", cfg.RedisURL)
		repo = redisRepo
		defer redisRepo.Close()
	}

	// ── 2. Reward client ─────────────────────────────────────────────────────
	rewardClient := client.NewRewardClient(cfg.RewardServiceURL)

	// ── 3. Usecase ───────────────────────────────────────────────────────────
	uc := usecase.New(repo, rewardClient)

	// ── 4. HTTP handler & router ─────────────────────────────────────────────
	h := handler.New(uc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("🏆 leaderboard service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down leaderboard service...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("leaderboard exited cleanly")
}
