package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port             int
	Env              string
	RedisURL         string
	RewardServiceURL string
	// RabbitMQ, where profile.updated arrives. Dialed with retry: the service
	// serves rankings while the broker is down.
	RabbitMQURL string
}

func Load() *Config {
	port := 8086
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("LEADERBOARD_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	redisURL := os.Getenv("LEADERBOARD_REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
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
		RabbitMQURL:      rabbitURL,
		Port:             port,
		Env:              env,
		RedisURL:         redisURL,
		RewardServiceURL: rewardURL,
	}
}
