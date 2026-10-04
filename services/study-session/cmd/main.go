package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/study-session/config"
	sessionAMQP "github.com/neennera/fishertimer/services/study-session/internal/adapter/amqp"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/client"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapters (Repository, Event Publisher, Timer Client)
	db := mustConnectDB(cfg)
	defer db.Close()
	repo := repository.NewPostgres(db)

	// RabbitMQ is dialed lazily by the publisher: the service starts even
	// while the broker is down, and events wait in the outbox meanwhile.
	publisher := sessionAMQP.NewPublisher(cfg.RabbitMQURL)
	defer publisher.Close()

	timerConn, err := grpc.NewClient(cfg.TimerGRPCTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("study-session: timer grpc client: %v", err)
	}
	defer timerConn.Close()
	timerClient := client.NewTimerClient(timerv1.NewStudyTimerServiceClient(timerConn))

	// 2. Inject into Application Usecase
	uc := usecase.New(repo)

	// 3. Background workers: outbox relay (RabbitMQ) and the room sweeper
	// (24h auto-end, disconnect timeout, idle timeout).
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()
	go usecase.NewOutboxRelay(repo, publisher, cfg.OutboxInterval).Run(workerCtx)
	go usecase.NewSweeper(uc, repo, timerClient, usecase.SweeperConfig{
		Interval:          cfg.SweepInterval,
		MaxSessionAge:     cfg.MaxSessionAge,
		DisconnectTimeout: cfg.DisconnectTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}).Run(workerCtx)

	// 4. Inject into Driving Adapters (HTTP and gRPC share the same usecase)
	h := handler.New(uc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	grpcServer := grpc.NewServer()
	sessionv1.RegisterStudySessionServiceServer(grpcServer, handler.NewGRPC(uc))
	if cfg.Env != "production" {
		// Lets tools such as grpcurl discover the RPCs without the .proto file.
		reflection.Register(grpcServer)
	}

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

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("grpc listen error: %s\n", err)
	}
	go func() {
		log.Printf("study-session gRPC listening on port %d", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc serve error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down study-session service...")
	stopWorkers()
	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("study-session exited cleanly")
}

// mustConnectDB opens session_db and refuses to start without it: room
// membership and the event outbox must survive a restart, or Study Timer
// would never hear about leaves that happened before it.
func mustConnectDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("study-session: cannot open session_db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("study-session: session_db unreachable (%v). Start it with `pnpm db:up` "+
			"or fix SESSION_DATABASE_URL (%s)", err, cfg.DatabaseURL)
	}

	log.Printf("study-session: connected to session_db")
	return db
}
