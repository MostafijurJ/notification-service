package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	customErr "github.com/mostafijurj/notification-service/internal/errors"
	"github.com/segmentio/kafka-go"
)

// ProduceMessage publishes a message to the given topic
func ProduceMessage(broker, topic, key, value string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer func(writer *kafka.Writer) {
		err := writer.Close()
		if err != nil {
			log.Printf("⚠️  Warning: failed to close Kafka writer: %v", err)
		}
	}(writer)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	}
	if err := writer.WriteMessages(ctx, msg); err != nil {
		return customErr.New("KAFKA_PRODUCE_FAILED", fmt.Sprintf("failed to produce message: %v", err))
	}
	log.Println("✅ Produced message to Kafka")
	return nil
}

// ConsumeMessage consumes a single message from the given topic and partition
func ConsumeMessage(broker, topic string, partition int) (string, string, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{broker},
		Topic:     topic,
		Partition: partition,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			log.Printf("⚠️  Warning: failed to close Kafka reader: %v", err)
		} else {
			log.Println("ℹ️  Kafka reader closed")
		}
	}(reader)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		return "", "", customErr.New("KAFKA_CONSUME_FAILED", fmt.Sprintf("failed to consume message: %v", err))
	}
	log.Printf("✅ Consumed message from Kafka: key=%s value=%s", string(msg.Key), string(msg.Value))
	return string(msg.Key), string(msg.Value), nil
}
