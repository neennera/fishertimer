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

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/neennera/fishertimer/services/reward/config"
	"github.com/neennera/fishertimer/services/reward/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/reward/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapter (Repository)
	mongoClient := mustConnectMongo(cfg)
	defer mongoClient.Disconnect(context.Background())
	repo := repository.NewMongo(mongoClient.Database("reward_db"))

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
		log.Printf("reward service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down reward service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("reward exited cleanly")
}

// mustConnectMongo opens reward_db and refuses to start without it: the
// catalog and every unlocked item live there, so running on no store at all
// would mean ViewRewards has nothing real to read.
func mustConnectMongo(cfg *config.Config) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoDBURI))
	if err != nil {
		log.Fatalf("reward: cannot connect to reward_db: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("reward: reward_db unreachable (%v). Start it with `pnpm db:up` "+
			"or fix REWARD_MONGODB_URI", err)
	}

	log.Printf("reward: connected to reward_db")
	return client
}
