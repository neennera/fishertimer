package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	GRPCPort        int
	Env             string
	DatabaseURL     string
	RabbitMQURL     string
	TimerGRPCTarget string

	// OutboxInterval is how often the relay publishes queued events.
	OutboxInterval time.Duration
	// SweepInterval is how often the time-based room rules run.
	SweepInterval time.Duration
	// MaxSessionAge auto-ends rooms (UC-03 E-6). 0 disables.
	MaxSessionAge time.Duration
	// DisconnectTimeout auto-leaves participants without heartbeats
	// (UC-03 E-3). 0 disables.
	DisconnectTimeout time.Duration
	// IdleTimeout auto-leaves participants with no running work cycle
	// (UC-03 E-7, IDLE_TIMEOUT_MINUTES). 0 disables.
	IdleTimeout time.Duration
}

func Load() *Config {
	port := 8083
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("SESSION_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("SESSION_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5433/session_db?sslmode=disable"
	}

	return &Config{
		Port:              port,
		GRPCPort:          envInt("SESSION_GRPC_PORT", 50052),
		Env:               env,
		DatabaseURL:       dbURL,
		RabbitMQURL:       envString("RABBITMQ_URL", "amqp://admin:adminpassword@localhost:5672/"),
		TimerGRPCTarget:   envString("TIMER_GRPC_TARGET", "localhost:50051"),
		OutboxInterval:    time.Duration(envInt("SESSION_OUTBOX_INTERVAL_MS", 500)) * time.Millisecond,
		SweepInterval:     time.Duration(envInt("SESSION_SWEEP_INTERVAL_SECONDS", 5)) * time.Second,
		MaxSessionAge:     time.Duration(envInt("SESSION_MAX_AGE_HOURS", 24)) * time.Hour,
		DisconnectTimeout: time.Duration(envInt("SESSION_DISCONNECT_TIMEOUT_SECONDS", 60)) * time.Second,
		IdleTimeout:       time.Duration(envInt("IDLE_TIMEOUT_MINUTES", 10)) * time.Minute,
	}
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
