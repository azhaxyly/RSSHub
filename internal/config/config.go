package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	TimerInterval time.Duration
	WorkersCount  int
	PGHost        string
	PGPort        string
	PGUser        string
	PGPassword    string
	PGDBName      string
	PGSSLmode     string
}

func LoadConfig() (*Config, error) {
	intervalStr := os.Getenv("CLI_APP_TIMER_INTERVAL")
	if intervalStr == "" {
		intervalStr = "3m" // Default
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, err
	}

	workersStr := os.Getenv("CLI_APP_WORKERS_COUNT")
	if workersStr == "" {
		workersStr = "3" // Default
	}
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		TimerInterval: interval,
		WorkersCount:  workers,
		PGHost:        os.Getenv("POSTGRES_HOST"),
		PGPort:        os.Getenv("POSTGRES_PORT"),
		PGUser:        os.Getenv("POSTGRES_USER"),
		PGPassword:    os.Getenv("POSTGRES_PASSWORD"),
		PGDBName:      os.Getenv("POSTGRES_DBNAME"),
		PGSSLmode:     "disable", // Default
	}, nil
}
