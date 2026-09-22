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

	"github.com/neennera/fishertimer/services/study-session/config"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/client"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapter (Repository & Collaborator Client)
	repo := repository.NewInMemory()
	rewardClient := client.NewRewardClient(cfg.RewardServiceURL)

	// 2. Inject into Application Usecase
	uc := usecase.New(repo, rewardClient)

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
		log.Printf("study-session service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down study-session service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("study-session exited cleanly")
}
