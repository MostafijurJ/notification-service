package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/db"
	ikafka "github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/mostafijurj/notification-service/internal/repository"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()
	dbConn, err := db.OpenGormPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	repo := repository.NewRepository(dbConn)
	log.Println("Worker inapp started")

	h := func(ctx context.Context, m kafkago.Message) error {
		id, _ := strconv.ParseInt(string(m.Value), 10, 64)
		// For demo, mark sent and insert delivery attempt
		_ = repo.UpdateNotificationStatus(ctx, id, "sent")
		_ = repo.InsertDeliveryAttempt(ctx, id, 1, nil, "success", nil, nil)
		return nil
	}

	if err := ikafka.ConsumeLoop(cfg.KafkaBrokers, ikafka.TopicReadyInAppLow, "worker-inapp-low", h); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
