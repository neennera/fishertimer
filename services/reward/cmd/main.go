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

	"github.com/neennera/fishertimer/services/reward/config"
	"github.com/neennera/fishertimer/services/reward/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/reward/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/reward/internal/domain"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

func main() {
	cfg := config.Load()

	// ── 1. Repository (MongoDB if available, in-memory fallback) ────────────
	var repo domain.Repository
	ctx := context.Background()

	mongoRepo, err := repository.NewMongo(ctx, cfg.MongoDBURI)
	if err != nil {
		log.Printf("⚠️  MongoDB unavailable (%v) — falling back to in-memory seed data", err)
		repo = repository.NewInMemory()
	} else {
		log.Printf("✅ Connected to MongoDB at %s", cfg.MongoDBURI)
		repo = mongoRepo
	}

	// ── 2. Usecase ───────────────────────────────────────────────────────────
	uc := usecase.New(repo)

	// ── 3. HTTP handler & router ─────────────────────────────────────────────
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
		log.Printf("🐟 reward service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down reward service...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("reward exited cleanly")
}
