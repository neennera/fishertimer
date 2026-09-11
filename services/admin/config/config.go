package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int
	Env  string
}

func Load() *Config {
	port := 8087
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	return &Config{
		Port: port,
		Env:  env,
	}
}
