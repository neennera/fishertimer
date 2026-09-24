package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  int
	Env                   string
	AccountServiceURL     string
	TimerServiceURL       string
	LeaderboardServiceURL string
	SessionServiceURL     string
	RewardServiceURL      string
}

func Load() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("API_GATEWAY_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	accountURL := os.Getenv("ACCOUNT_SERVICE_URL")
	if accountURL == "" {
		accountURL = "http://localhost:8082"
	}

	timerURL := os.Getenv("TIMER_SERVICE_URL")
	if timerURL == "" {
		timerURL = "http://localhost:8084"
	}

	leaderboardURL := os.Getenv("LEADERBOARD_SERVICE_URL")
	if leaderboardURL == "" {
		leaderboardURL = "http://localhost:8086"
	}

	sessionURL := os.Getenv("SESSION_SERVICE_URL")
	if sessionURL == "" {
		sessionURL = "http://localhost:8083"
	}

	rewardURL := os.Getenv("REWARD_SERVICE_URL")
	if rewardURL == "" {
		rewardURL = "http://localhost:8085"
	}

	return &Config{
		Port:                  port,
		Env:                   env,
		AccountServiceURL:     accountURL,
		TimerServiceURL:       timerURL,
		LeaderboardServiceURL: leaderboardURL,
		SessionServiceURL:     sessionURL,
		RewardServiceURL:      rewardURL,
	}
}
