package config

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
	AppEnv      string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		AppEnv:      os.Getenv("APP_ENV"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// default to 50051 — standard gRPC port convention
	if cfg.ServerPort == "" {
		cfg.ServerPort = "50051"
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	return cfg, nil
}