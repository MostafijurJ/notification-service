package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type HandlerFunc func(ctx context.Context, m kafka.Message) error

func ConsumeLoop(broker, topic, groupID string, handler HandlerFunc) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer func() {
		_ = r.Close()
		log.Printf("Kafka reader for topic=%s group=%s closed", topic, groupID)
	}()

	ctx := context.Background()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := handler(ctx, m); err != nil {
			log.Printf("Handler error: %v", err)
		}
	}
}