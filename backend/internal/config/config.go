package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://reload:reload@localhost:5432/reloaddb?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
