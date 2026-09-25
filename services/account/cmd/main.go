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

	_ "github.com/lib/pq"

	"github.com/neennera/fishertimer/services/account/config"
	"github.com/neennera/fishertimer/services/account/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/account/internal/adapter/oauth"
	"github.com/neennera/fishertimer/services/account/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/account/internal/adapter/token"
	"github.com/neennera/fishertimer/services/account/internal/usecase"
)

func main() {
	cfg := config.Load()

	// 1. Driven adapters: database, Google, session tokens.
	db := mustConnectDB(cfg)
	defer db.Close()
	repo := repository.NewPostgres(db)

	provider := oauth.New(oauth.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		AuthURL:      cfg.GoogleAuthURL,
		TokenURL:     cfg.GoogleTokenURL,
		UserInfoURL:  cfg.GoogleUserInfoURL,
	})

	tokens := token.New(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiry)

	// 2. Application usecase.
	uc := usecase.New(repo, provider, tokens)

	// 3. Driving adapter (HTTP).
	h := handler.New(uc, handler.Options{
		FrontendURL:     cfg.FrontendURL,
		CookieSecure:    cfg.CookieSecure,
		SessionMaxAge:   cfg.JWTExpiry,
		OAuthConfigured: cfg.OAuthConfigured(),
	})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("account service listening on port %d [%s]", cfg.Port, cfg.Env)
		if cfg.OAuthConfigured() {
			log.Printf("  -> google oauth redirect_uri: %s", cfg.GoogleRedirectURL)
		} else {
			log.Printf("  !! GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET / GOOGLE_REDIRECT_URI missing: sign-in returns 503")
		}
		if cfg.JWTSecret == "" {
			log.Printf("  !! JWT_SECRET missing: session tokens cannot be issued")
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down account service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %s\n", err)
	}
	log.Printf("account exited cleanly")
}

// mustConnectDB opens account_db and refuses to start without it. Accounts are
// the service's whole job: running with no durable store would hand out
// sessions for users that disappear on the next restart.
func mustConnectDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("account: cannot open account_db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("account: account_db unreachable (%v). Start it with `pnpm db:up` "+
			"or fix ACCOUNT_DATABASE_URL (%s)", err, cfg.DatabaseURL)
	}

	log.Printf("account: connected to account_db")
	return db
}
