package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neennera/fishertimer/pkg/discovery"
	"github.com/neennera/fishertimer/services/admin/config"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/client"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/middleware"
	"github.com/neennera/fishertimer/services/admin/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	// Service discovery (DISCOVERY_MODE=consul): peers are looked up in the
	// registry, falling back to the configured addresses. In the default static
	// mode it simply returns those addresses.
	disc := discovery.FromEnv(map[string]string{
		discovery.StudySession: cfg.SessionServiceURL,
	})

	if cfg.JWTSecret == "" {
		log.Printf("admin: JWT_SECRET is not set; every admin request will be rejected with 401")
	}

	// 1. Instantiate Driven Adapters: live room state comes from Study
	// Session, the moderation audit log lives in admin_db.
	db := mustConnectDB(cfg)
	defer db.Close()
	sessions := repository.NewSessionService(cfg.SessionServiceURL, disc.Locator.HTTPClient(discovery.StudySession))
	auditLogs := repository.NewPostgres(db)
	sessionClient := client.NewSessionClient(cfg.SessionServiceURL, disc.Locator.HTTPClient(discovery.StudySession))

	// 2. Inject into Application Usecase
	uc := usecase.New(sessions, auditLogs, sessionClient)

	// 3. Inject into Driving Adapter (HTTP Handler, admin-only)
	h := handler.New(uc, middleware.NewVerifier(cfg.JWTSecret, cfg.JWTIssuer))

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
		log.Printf("admin service listening on port %d [%s]", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	// Announce this service to the registry (DISCOVERY_MODE=consul) so peers can
	// find it. Does nothing in the default static mode.
	stopAnnounce := disc.Announce(disc.HTTP(discovery.Admin, cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stopAnnounce() // leave the registry first, so peers stop sending requests here

	log.Printf("Shutting down admin service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("admin exited cleanly")
}

// mustConnectDB opens admin_db and refuses to start without it: every
// moderation action must be audited (UC-10), so running without the audit
// store would either block all moderation or lose the trail.
func mustConnectDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("admin: cannot open admin_db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("admin: admin_db unreachable (%v). Start it with `pnpm db:up` "+
			"or fix ADMIN_DATABASE_URL (%s)", err, cfg.DatabaseURL)
	}

	log.Printf("admin: connected to admin_db")
	return db
}
