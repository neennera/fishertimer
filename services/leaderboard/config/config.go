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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	rewardURL := os.Getenv("REWARD_SERVICE_URL")
	if rewardURL == "" {
		rewardURL = "http://localhost:8085"
	}

	return &Config{
		Port:             port,
		Env:              env,
		RedisURL:         redisURL,
		RewardServiceURL: rewardURL,
	}
}
