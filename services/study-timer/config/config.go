package config

import (
	"os"
	"strconv"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

type Config struct {
	Port              int
	GRPCPort          int
	Env               string
	DatabaseURL       string
	RewardServiceURL  string
	RabbitMQURL       string
	SessionGRPCTarget string
	// SweepInterval is how often due cycles are completed, long pauses
	// discarded and undelivered rewards retried.
	SweepInterval time.Duration
	Limits        domain.Limits
}

func Load() *Config {
	port := 8084
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("TIMER_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("TIMER_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgrespassword@localhost:5434/timer_db?sslmode=disable"
	}

	d := domain.DefaultLimits
	return &Config{
		Port:              port,
		GRPCPort:          envInt("TIMER_GRPC_PORT", 50051),
		Env:               env,
		DatabaseURL:       dbURL,
		RewardServiceURL:  envString("REWARD_SERVICE_URL", "http://localhost:8085"),
		RabbitMQURL:       envString("RABBITMQ_URL", "amqp://admin:adminpassword@localhost:5672/"),
		SessionGRPCTarget: envString("SESSION_GRPC_TARGET", "localhost:50052"),
		SweepInterval:     time.Duration(envInt("TIMER_SWEEP_INTERVAL_SECONDS", 2)) * time.Second,
		Limits: domain.Limits{
			MinWorkMinutes:  envInt("MIN_WORK_MINUTES", d.MinWorkMinutes),
			MaxWorkMinutes:  envInt("MAX_WORK_MINUTES", d.MaxWorkMinutes),
			MinRestMinutes:  envInt("MIN_REST_MINUTES", d.MinRestMinutes),
			MaxRestMinutes:  envInt("MAX_REST_MINUTES", d.MaxRestMinutes),
			MaxPauseMinutes: envInt("MAX_PAUSE_MINUTES", d.MaxPauseMinutes),
		},
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
