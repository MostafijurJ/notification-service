package utils

import (
	"database/sql"
	"fmt"
	"log"
	_ "time"

	"github.com/segmentio/kafka-go"
)

// TestPostgresConnection checks if PostgreSQL is reachable
func TestPostgresConnection(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Postgres connection failed: %v", err)
	}
	log.Println("✅ PostgreSQL connection successful")
	return nil
}

// TestKafkaConnection checks if Kafka broker is reachable
func TestKafkaConnection(broker string, topic string) error {
	conn, err := kafka.DialLeader(nil, "tcp", broker, topic, 0)
	if err != nil {
		return fmt.Errorf("Kafka connection failed: %v", err)
	}
	defer conn.Close()

	log.Println("✅ Kafka broker and topic reachable")
	return nil
}

// InitKafkaTopic creates the topic if not exists (optional)
func InitKafkaTopic(broker, topic string, partitions int) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("Failed to connect Kafka broker: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("Failed to get Kafka controller: %v", err)
	}

	ctrlConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("Failed to connect Kafka controller: %v", err)
	}
	defer ctrlConn.Close()

	topicConfigs := []kafka.TopicConfig{{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	}}

	err = ctrlConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("Failed to create topic: %v", err)
	}

	log.Printf("✅ Kafka topic '%s' initialized", topic)
	return nil
}
