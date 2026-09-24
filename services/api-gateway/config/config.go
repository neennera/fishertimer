package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	port := 8000
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
