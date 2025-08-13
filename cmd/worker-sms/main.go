package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/mostafijurj/notification-service/config"
	ikafka "github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/mostafijurj/notification-service/internal/repository"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil { log.Fatalf("db open: %v", err) }
	repo := repository.NewRepository(db)
	log.Println("Worker sms started")
	h := func(ctx context.Context, m kafkago.Message) error {
		id, _ := strconv.ParseInt(string(m.Value), 10, 64)
		_ = repo.UpdateNotificationStatus(ctx, id, "sent")
		_ = repo.InsertDeliveryAttempt(ctx, id, 1, nil, "success", nil, nil)
		return nil
	}
	_ = ikafka.ConsumeLoop(cfg.KafkaBrokers, ikafka.TopicReadySMSLow, "worker-sms-low", h)
}