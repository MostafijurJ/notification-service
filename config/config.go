package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	PostgresDSN  string
	KafkaBrokers string
}

func Load() *Config {
	// Determine active environment (default: dev)
	env := getEnv("APP_ENV", "dev")
	envFile := fmt.Sprintf("config/%s.env", env)

	// Load corresponding .env file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: could not load %s, using default env vars", envFile)
	}

	cfg := &Config{
		AppEnv:       env,
		AppPort:      getEnv("APP_PORT", "8081"),
		PostgresDSN:  getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:19092"),
	}
	log.Printf("Loaded configuration for %s environment", env)
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
