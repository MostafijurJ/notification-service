package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/db"
	"github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/mostafijurj/notification-service/internal/models"
)

func main() {
	cfg := config.Load()
	dbConn, err := db.OpenGormPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	ctx := context.Background()
	log.Println("Scheduler started")

	for {
		// find due notifications using GORM
		var notifications []models.Notification
		err := dbConn.WithContext(ctx).
			Where("status = ? AND scheduled_at <= ?", "scheduled", time.Now()).
			Limit(100).
			Find(&notifications).Error

		if err != nil {
			log.Printf("query: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, notification := range notifications {
			ready := resolveReadyTopic(notification.Channel, notification.Priority)
			_ = kafka.ProduceMessage(cfg.KafkaBrokers, ready, "", strconv.FormatInt(notification.ID, 10))

			// Update status using GORM
			_ = dbConn.WithContext(ctx).
				Model(&models.Notification{}).
				Where("id = ?", notification.ID).
				Update("status", "enqueued")
		}

		time.Sleep(1 * time.Second)
	}
}

func resolveReadyTopic(ch, pr string) string {
	switch ch {
	case "email":
		if pr == "high" {
			return kafka.TopicReadyEmailHigh
		}
		return kafka.TopicReadyEmailLow
	case "sms":
		if pr == "high" {
			return kafka.TopicReadySMSHigh
		}
		return kafka.TopicReadySMSLow
	case "push":
		if pr == "high" {
			return kafka.TopicReadyPushHigh
		}
		return kafka.TopicReadyPushLow
	case "inapp":
		if pr == "high" {
			return kafka.TopicReadyInAppHigh
		}
		return kafka.TopicReadyInAppLow
	default:
		return kafka.TopicReadyInAppLow
	}
}
