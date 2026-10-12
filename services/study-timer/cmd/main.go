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
	"github.com/neennera/fishertimer/services/study-timer/config"
	timerAMQP "github.com/neennera/fishertimer/services/study-timer/internal/adapter/amqp"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/client"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapters (Repository, Reward and Study Session clients)
	db := mustConnectDB(cfg)
	defer db.Close()
	repo := repository.NewPostgres(db)
	rewardClient := client.NewRewardClient(cfg.RewardServiceURL)

	// Participant counts for AwardReward come from Study Session over gRPC.
	// The connection is lazy: rewards wait as PENDING while it is down.
	sessionConn, err := grpc.NewClient(cfg.SessionGRPCTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("study-timer: session grpc client: %v", err)
	}
	defer sessionConn.Close()
	sessionClient := client.NewSessionClient(sessionv1.NewStudySessionServiceClient(sessionConn))

	// 2. Inject into Application Usecase
	uc := usecase.New(repo, rewardClient,
		usecase.WithSessionClient(sessionClient),
		usecase.WithLimits(cfg.Limits))

	// 3. Inject into Driving Adapters (HTTP and gRPC share the same usecase)
	h := handler.New(uc)

	grpcServer := grpc.NewServer()
	timerv1.RegisterStudyTimerServiceServer(grpcServer, handler.NewGRPC(uc))
	if cfg.Env != "production" {
		// Lets tools such as grpcurl discover the RPCs without the .proto file.
		reflection.Register(grpcServer)
	}

	// 4. Background workers: the RabbitMQ consumer (reconnects on its own,
	// events wait in the durable queue meanwhile) and the sweeper, which
	// completes cycles whose time is up even with no client watching,
	// discards cycles paused too long and retries undelivered rewards.
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()
	go timerAMQP.NewConsumer(cfg.RabbitMQURL, uc).Run(workerCtx)
	go func() {
		ticker := time.NewTicker(cfg.SweepInterval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				uc.Sweep(workerCtx)
			}
		}
	}()

	// 5. Setup Router & Server
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("study-timer service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("grpc listen error: %s\n", err)
	}
	go func() {
		log.Printf("study-timer gRPC listening on port %d", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc serve error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down study-timer service...")
	stopWorkers()
	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("study-timer exited cleanly")
}

// mustConnectDB opens timer_db and refuses to start without it: statistics
// (sessions joined, cycles completed, total focus time) are computed with
// SUM/COUNT queries against real rows, so running on an ephemeral store would
// make every user's stats reset on restart.
func mustConnectDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("study-timer: cannot open timer_db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("study-timer: timer_db unreachable (%v). Start it with `pnpm db:up` "+
			"or fix TIMER_DATABASE_URL (%s)", err, cfg.DatabaseURL)
	}

	log.Printf("study-timer: connected to timer_db")
	return db
}
