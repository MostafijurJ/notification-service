package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestRunMigrations_Smoke(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil { t.Fatalf("open: %v", err) }
	defer db.Close()
	if err := RunMigrations(context.Background(), db); err != nil { t.Fatalf("migrate: %v", err) }
}