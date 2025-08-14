package main

import (
	"context"
	"log"
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
	log.Println("Worker push started")

	h := func(ctx context.Context, m kafkago.Message) error {
		id, _ := strconv.ParseInt(string(m.Value), 10, 64)
		_ = repo.UpdateNotificationStatus(ctx, id, "sent")
		_ = repo.InsertDeliveryAttempt(ctx, id, 1, nil, "success", nil, nil)
		return nil
	}

	_ = ikafka.ConsumeLoop(cfg.KafkaBrokers, ikafka.TopicReadyPushLow, "worker-push-low", h)
}
