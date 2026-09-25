package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port        int
	Env         string
	DatabaseURL string

	// Google OAuth 2.0 client credentials (Google Cloud Console -> Credentials).
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Google endpoints, overridable so tests can point at a stub server.
	GoogleAuthURL     string
	GoogleTokenURL    string
	GoogleUserInfoURL string

	// Session token (UC-06: a 7-day JWT carrying user_id and role).
	JWTSecret string
	JWTIssuer string
	JWTExpiry time.Duration

	// Where the browser is sent once the OAuth round trip finishes.
	FrontendURL  string
	CookieSecure bool
}

// OAuthConfigured reports whether real Google credentials are present. Without
// them the service still boots (health checks stay green) but the sign-in
// endpoints answer 503 instead of redirecting users to a broken consent screen.
func (c *Config) OAuthConfigured() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != "" && c.GoogleRedirectURL != ""
}

// loadEnv pulls the monorepo root .env into the process environment so that
// `pnpm dev` / `go run` behave like a container with env injected. Real
// environment variables always win.
func loadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		envPath := filepath.Join(dir, ".env")
		if file, err := os.Open(envPath); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					k := strings.TrimSpace(parts[0])
					v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
					if os.Getenv(k) == "" {
						os.Setenv(k, v)
					}
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

func Load() *Config {
	loadEnv()

	port := 8082
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("ACCOUNT_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	env := getenv("ENV", "development")

	dbURL := os.Getenv("ACCOUNT_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5432/account_db?sslmode=disable"
	}

	// UC-06 specifies a 7-day session token.
	expiryHours := 168
	if h := os.Getenv("JWT_EXPIRATION_HOURS"); h != "" {
		if val, err := strconv.Atoi(h); err == nil && val > 0 {
			expiryHours = val
		}
	}

	frontendURL := getenv("NEXT_PUBLIC_APP_URL", "http://localhost:3000")

	return &Config{
		Port:        port,
		Env:         env,
		DatabaseURL: dbURL,

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URI", frontendURL+"/api/auth/google/callback"),

		GoogleAuthURL:     getenv("GOOGLE_AUTH_URL", "https://accounts.google.com/o/oauth2/v2/auth"),
		GoogleTokenURL:    getenv("GOOGLE_TOKEN_URL", "https://oauth2.googleapis.com/token"),
		GoogleUserInfoURL: getenv("GOOGLE_USERINFO_URL", "https://openidconnect.googleapis.com/v1/userinfo"),

		JWTSecret: os.Getenv("JWT_SECRET"),
		JWTIssuer: getenv("JWT_ISSUER", "fishertimer-account"),
		JWTExpiry: time.Duration(expiryHours) * time.Hour,

		FrontendURL:  frontendURL,
		CookieSecure: env == "production",
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
