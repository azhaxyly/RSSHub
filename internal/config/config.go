package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL  string
	TimeInterval time.Duration
  NumWorkers int
	DSN 				string
	LogLevel     string
}

func LoadConfig() (*Config, error) {
	timeInterval, err := time.ParseDuration(getOrDefault("TIME_INTERVAL", "3m"))
	if err != nil {
		return nil, err
	}

	return &Config{
		TimeInterval: timeInterval,
		NumWorkers: getOrDefaultInt("NUM_WORKERS", 3),
		DSN: getOrDefault("DATABASE_URL", "postgres://rsshub:rsshub@localhost:5432/rsshub?sslmode=disable"),
		LogLevel:     getOrDefault("LOG_LEVEL", "development"),
	}, nil
}


func getOrDefault(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getOrDefaultInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	strc, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return strc
}


