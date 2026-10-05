package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port             int
	GRPCPort         int
	Env              string
	DatabaseURL      string
	RewardServiceURL string
	RabbitMQURL      string
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
	grpcPort := 50051
	if p := os.Getenv("TIMER_GRPC_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			grpcPort = val
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

	rewardURL := os.Getenv("REWARD_SERVICE_URL")
	if rewardURL == "" {
		rewardURL = "http://localhost:8085"
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://admin:adminpassword@localhost:5672/"
	}

	return &Config{
		Port:             port,
		GRPCPort:         grpcPort,
		Env:              env,
		DatabaseURL:      dbURL,
		RewardServiceURL: rewardURL,
		RabbitMQURL:      rabbitURL,
	}
}
