package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/mostafijurj/notification-service/internal/repository"
)

func main() {
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil { log.Fatalf("db open: %v", err) }
	repo := repository.NewRepository(db)
	log.Println("Worker inapp started")
	h := func(ctx context.Context, m kafka.Message) error {
		id, _ := strconv.ParseInt(string(m.Value), 10, 64)
		// For demo, mark sent and insert delivery attempt
		_ = repo.UpdateNotificationStatus(ctx, id, "sent")
		_ = repo.InsertDeliveryAttempt(ctx, id, 1, nil, "success", nil, nil)
		return nil
	}
	if err := kafka.ConsumeLoop(cfg.KafkaBrokers, kafka.TopicReadyInAppLow, "worker-inapp-low", h); err != nil {
		log.Println(err); os.Exit(1)
	}
}