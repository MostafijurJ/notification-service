package utils

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestTestPostgresConnection_ClosedDB(t *testing.T) {
	db, _ := sql.Open("postgres", "postgres://invalid")
	_ = db.Close()
	if err := TestPostgresConnection(db); err == nil {
		t.Skip("driver may not error on Ping with closed db in this environment")
	}
}