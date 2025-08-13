package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	PostgresDSN  string
	KafkaBrokers string
	RedisURL     string
}

func Load() *Config {
	// Determine active environment (default: dev)
	env := getEnv("APP_ENV", "dev")
	envFile := fmt.Sprintf("config/%s.env", env)

	// Load corresponding .env file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: could not load %s, using default env vars", envFile)
	}

	pgDSN := getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable")
	// Prefer JDBC_POSTGRES_URL if provided; normalize to lib/pq-compatible DSN
	if jdbcPg, ok := os.LookupEnv("POSTGRES_DSN"); ok && jdbcPg != "" {
		pgDSN = ensurePostgresSSLParam(normalizeJDBCURL(jdbcPg))
	} else if strings.HasPrefix(pgDSN, "jdbc:") {
		pgDSN = ensurePostgresSSLParam(normalizeJDBCURL(pgDSN))
	}

	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	// Support JDBC-style Redis URL too
	if jdbcRedis, ok := os.LookupEnv("JDBC_REDIS_URL"); ok && jdbcRedis != "" {
		redisURL = normalizeJDBCURL(jdbcRedis)
	} else if strings.HasPrefix(redisURL, "jdbc:") {
		redisURL = normalizeJDBCURL(redisURL)
	}

	cfg := &Config{
		AppEnv:       env,
		AppPort:      getEnv("APP_PORT", "8081"),
		PostgresDSN:  pgDSN,
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:19092"),
		RedisURL:     redisURL,
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

// normalizeJDBCURL strips the optional "jdbc:" prefix for URLs provided in JDBC format
func normalizeJDBCURL(raw string) string {
	if strings.HasPrefix(raw, "jdbc:") {
		return strings.TrimPrefix(raw, "jdbc:")
	}
	return raw
}

// ensurePostgresSSLParam ensures sslmode is set for postgres/postgresql URL DSNs when not provided
func ensurePostgresSSLParam(dsn string) string {
	if !(strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://")) {
		return dsn
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	q := u.Query()
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "disable")
		u.RawQuery = q.Encode()
		return u.String()
	}
	return dsn
}
