// Package config loads configurations from .env
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL        string
	ClientID     string
	ClientSecret string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DBURL:        os.Getenv("DATABASE_URL"),
		ClientID:     os.Getenv("OPEN_SKY_CLIENT_ID"),
		ClientSecret: os.Getenv("OPEN_SKY_CLIENT_SECRET"),
	}
}
