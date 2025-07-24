package main

import (
	"database/sql"
	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/utils"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	// PostgresSQL Connection Test
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect Postgres: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Warning: failed to close Postgres connection: %v", err)
		} else {
			log.Println("✅ Postgres connection closed successfully")
		}
	}(db)

	if err := utils.TestPostgresConnection(db); err != nil {
		log.Fatal(err)
	}

	// Kafka Connection Test
	topic := "notifications"
	if err := utils.InitKafkaTopic(cfg.KafkaBrokers, topic, 1); err != nil {
		log.Fatalf("❌ Kafka topic initialization failed: %v", err)
	}
	if err := utils.TestKafkaConnection(cfg.KafkaBrokers, topic); err != nil {
		log.Fatal(err)
	}

	log.Println("🚀 All connections are healthy. Ready to start the service.")
}
