package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	kafkaPkg "github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/redis/go-redis/v9"
	
)

// TestPostgresConnection checks if PostgreSQL is reachable
func TestPostgresConnection(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL connection failed: %v", err)
	}
	log.Println("✅ PostgreSQL connection successful")
	return nil
}

// TestKafkaConnection checks if Kafka broker is reachable
func TestKafkaConnection(broker string, topic string) error {
	// Use context with timeout (to avoid panic with nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialLeader(ctx, "tcp", broker, topic, 0)
	if err != nil {
		return fmt.Errorf("kafka connection failed: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("⚠️  Warning: failed to close Kafka connection: %v", err)
		} else {
			log.Println("ℹ️  Kafka connection closed")
		}
	}()

	log.Println("✅ Kafka broker and topic reachable")
	return nil
}

// InitKafkaTopic creates the topic if it does not exist
func InitKafkaTopic(broker, topic string, partitions int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect Kafka broker: %v", err)
	}
	log.Println("✅ Connected to Kafka broker")
	log.Printf("Ensuring Kafka topic '%s' with %d partitions", topic, partitions)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("⚠️  Warning: failed to close Kafka connection: %v", err)
		} else {
			log.Println("ℹ️  Kafka connection closed")
		}
	}()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get Kafka controller: %v", err)
	}
	controllerConn, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka controller: %v", err)
	}
	defer func() {
		if err := controllerConn.Close(); err != nil {
			log.Printf("⚠️  Warning: failed to close Kafka controller connection: %v", err)
		}
	}()

	topicConfigs := []kafka.TopicConfig{{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	}}
	if err := controllerConn.CreateTopics(topicConfigs...); err != nil {
		return fmt.Errorf("failed to create Kafka topic: %v", err)
	}
	log.Println("✅ Kafka topic ensured/created")

	return nil
}

// TestKafkaProduceConsume checks Kafka by producing and consuming a test message
func TestKafkaProduceConsume(broker, topic string) error {
	testKey := "healthcheck-key"
	testValue := "healthcheck-value"

	if err := kafkaPkg.ProduceMessage(broker, topic, testKey, testValue); err != nil {
		return fmt.Errorf("produce failed: %v", err)
	}
	key, value, err := kafkaPkg.ConsumeMessage(broker, topic, 0)
	if err != nil {
		return fmt.Errorf("consume failed: %v", err)
	}
	if key != testKey || value != testValue {
		return fmt.Errorf("mismatched message: got key=%s value=%s", key, value)
	}
	log.Println("✅ Kafka produce/consume healthcheck passed")
	return nil
}

// TestRedisConnection checks if Redis is reachable
func TestRedisConnection(redisURL string) error {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	defer func(client *redis.Client) {
		if err := client.Close(); err != nil {
			log.Printf("⚠️  Warning: failed to close Redis connection: %v", err)
		} else {
			log.Println("ℹ️  Redis connection closed")
		}
	}(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection with PING command
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis connection failed: %v", err)
	}

	log.Println("✅ Redis connection successful")
	return nil
}
