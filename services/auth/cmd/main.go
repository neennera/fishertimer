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

	"github.com/neennera/fishertimer/services/auth/config"
	"github.com/neennera/fishertimer/services/auth/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/auth/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/auth/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapter (Repository)
	repo := repository.NewInMemory()

	// 2. Inject into Application Usecase
	uc := usecase.New(repo)

	// 3. Inject into Driving Adapter (HTTP Handler)
	h := handler.New(uc)

	// 4. Setup Router & Server
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("auth service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down auth service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("auth exited cleanly")
}
