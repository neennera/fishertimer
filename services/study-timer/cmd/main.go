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
	"google.golang.org/grpc/reflection"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/study-timer/config"
	timerAMQP "github.com/neennera/fishertimer/services/study-timer/internal/adapter/amqp"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/client"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cfg := config.Load()

	// 1. Instantiate Driven Adapter (Repository & Reward Collaborator Client)
	db := mustConnectDB(cfg)
	defer db.Close()
	repo := repository.NewPostgres(db)
	rewardClient := client.NewRewardClient(cfg.RewardServiceURL)

	// 2. Inject into Application Usecase
	uc := usecase.New(repo, rewardClient)

	// 3. Inject into Driving Adapters (HTTP and gRPC share the same usecase)
	h := handler.New(uc)

	grpcServer := grpc.NewServer()
	timerv1.RegisterStudyTimerServiceServer(grpcServer, handler.NewGRPC(uc))
	if cfg.Env != "production" {
		// Lets tools such as grpcurl discover the RPCs without the .proto file.
		reflection.Register(grpcServer)
	}

	// 4. Setup RabbitMQ Connection, Topology & Consumer Worker
	amqpConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("study-timer: warning: rabbitmq unreachable (%v) at %s. Running without message queue", err, cfg.RabbitMQURL)
	} else {
		defer amqpConn.Close()
		log.Printf("study-timer: connected to RabbitMQ at %s", cfg.RabbitMQURL)

		amqpCh, err := amqpConn.Channel()
		if err != nil {
			log.Fatalf("study-timer: cannot open rabbitmq channel: %v", err)
		}
		defer amqpCh.Close()

		if err := timerAMQP.DeclareTopology(amqpCh); err != nil {
			log.Fatalf("study-timer: cannot declare rabbitmq topology: %v", err)
		}

		consumerCtx, cancelConsumer := context.WithCancel(context.Background())
		defer cancelConsumer()
		consumer := timerAMQP.NewConsumer(amqpCh, uc)
		go func() {
			if err := consumer.Start(consumerCtx); err != nil {
				log.Printf("study-timer: consumer error: %v", err)
			}
		}()
	}

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
