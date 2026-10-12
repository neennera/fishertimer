package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	Env               string
	DatabaseURL       string
	SessionServiceURL string
	JWTSecret         string
	JWTIssuer         string
}

// loadEnv pulls the monorepo root .env into the process environment so the
// admin service verifies session JWTs with the same JWT_SECRET the account
// service signs with. Variables already set in the environment win.
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

	port := 8087
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("ADMIN_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("ADMIN_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5435/admin_db?sslmode=disable"
	}

	sessionURL := os.Getenv("SESSION_SERVICE_URL")
	if sessionURL == "" {
		sessionURL = "http://localhost:8083"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "fishertimer-account"
	}

	return &Config{
		Port:              port,
		Env:               env,
		DatabaseURL:       dbURL,
		SessionServiceURL: sessionURL,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTIssuer:         jwtIssuer,
	}
}
