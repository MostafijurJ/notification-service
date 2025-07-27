package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/logger"
	"github.com/mostafijurj/notification-service/internal/middleware"
	"github.com/mostafijurj/notification-service/internal/routes"
	"github.com/mostafijurj/notification-service/internal/utils"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	logger.Init()

	preConnectionTest(nil, cfg)

	router := mux.NewRouter()
	router.Use(middleware.RequestIDMiddleware)
	routes.HomeRoutes(router)

	logger.Info.Printf("Server starting on :%s", cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func preConnectionTest(err error, cfg *config.Config) {
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

	// Produce and consume a test message
	err = utils.TestKafkaProduceConsume(cfg.KafkaBrokers, topic)
	if err != nil {
		log.Fatalf("❌ Kafka produce/consume test failed: %v", err)
	}
	log.Println("🚀 All connections are healthy. Ready to start the service.")
}
