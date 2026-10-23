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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/neennera/fishertimer/pkg/discovery"
	"github.com/neennera/fishertimer/pkg/discovery/grpcresolver"
	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/api-gateway/config"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

func main() {
	cfg := config.Load()

	// Service discovery (DISCOVERY_MODE=consul): peers are looked up in the
	// registry, falling back to the configured addresses. In the default static
	// mode it simply returns those addresses.
	disc := discovery.FromEnv(map[string]string{
		discovery.Account:                      cfg.AccountServiceURL,
		discovery.Leaderboard:                  cfg.LeaderboardServiceURL,
		discovery.Reward:                       cfg.RewardServiceURL,
		discovery.GRPC(discovery.StudyTimer):   cfg.TimerGRPCTarget,
		discovery.GRPC(discovery.StudySession): cfg.SessionGRPCTarget,
	})

	// 1. Connect to Study Timer over gRPC. The connection is lazy: the gateway
	// starts even if the timer is down, and timer routes answer 503 until it
	// is reachable. Services talk plaintext inside the private network.
	timerTarget, timerOpts := grpcresolver.Dial(disc.Locator, discovery.GRPC(discovery.StudyTimer), cfg.TimerGRPCTarget)
	timerConn, err := grpc.NewClient(timerTarget, append(timerOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
	if err != nil {
		log.Fatalf("timer grpc client: %s\n", err)
	}
	defer timerConn.Close()

	// Study Session is reached over gRPC the same way.
	sessionTarget, sessionOpts := grpcresolver.Dial(disc.Locator, discovery.GRPC(discovery.StudySession), cfg.SessionGRPCTarget)
	sessionConn, err := grpc.NewClient(sessionTarget, append(sessionOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
	if err != nil {
		log.Fatalf("session grpc client: %s\n", err)
	}
	defer sessionConn.Close()

	// 2. Instantiate Gateway Handler
	h := handler.New(cfg, timerv1.NewStudyTimerServiceClient(timerConn), sessionv1.NewStudySessionServiceClient(sessionConn), handler.WithDiscovery(disc.Locator))

	// 3. Setup Router & Server
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
		log.Printf("  -> Timer:       %s (gRPC)", cfg.TimerGRPCTarget)
		log.Printf("  -> Leaderboard: %s", cfg.LeaderboardServiceURL)
		log.Printf("  -> Session:     %s (gRPC)", cfg.SessionGRPCTarget)
		log.Printf("  -> Reward:      %s", cfg.RewardServiceURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	// Announce this service to the registry (DISCOVERY_MODE=consul) so peers can
	// find it. Does nothing in the default static mode.
	stopAnnounce := disc.Announce(disc.HTTP(discovery.APIGateway, cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stopAnnounce() // leave the registry first, so peers stop sending requests here

	log.Printf("Shutting down api-gateway service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("api-gateway exited cleanly")
}
