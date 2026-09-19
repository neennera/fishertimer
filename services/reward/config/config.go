package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port       int
	Env        string
	MongoDBURI string
}

func Load() *Config {
	port := 8085
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	} else if p := os.Getenv("REWARD_SERVICE_PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	mongoURI := os.Getenv("REWARD_MONGODB_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGODB_URI")
	}
	if mongoURI == "" {
		mongoURI = "mongodb://mongoadmin:mongopassword@localhost:27017/reward_db?authSource=admin"
	}

	return &Config{
		Port:       port,
		Env:        env,
		MongoDBURI: mongoURI,
	}
}
