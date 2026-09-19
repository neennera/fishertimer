package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Env         string
	DatabaseURL string
}

func Load() *Config {
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

	return &Config{
		Port:        port,
		Env:         env,
		DatabaseURL: dbURL,
	}
}
