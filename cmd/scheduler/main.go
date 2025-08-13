package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/mostafijurj/notification-service/config"
	"github.com/mostafijurj/notification-service/internal/kafka"
)

func main() {
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil { log.Fatalf("db open: %v", err) }
	ctx := context.Background()
	log.Println("Scheduler started")
	for {
		// find due notifications
		rows, err := db.QueryContext(ctx, `SELECT id, channel, priority FROM notifications WHERE status='scheduled' AND scheduled_at <= NOW() LIMIT 100`)
		if err != nil { log.Printf("query: %v", err); time.Sleep(2*time.Second); continue }
		var ids []struct{ id int64; ch, pr string }
		for rows.Next() { var id int64; var ch, pr string; _ = rows.Scan(&id, &ch, &pr); ids = append(ids, struct{ id int64; ch, pr string }{id, ch, pr}) }
		_ = rows.Close()
		for _, it := range ids {
			ready := resolveReadyTopic(it.ch, it.pr)
			_ = kafka.ProduceMessage(cfg.KafkaBrokers, ready, "", strconv.FormatInt(it.id, 10))
			_, _ = db.ExecContext(ctx, `UPDATE notifications SET status='enqueued' WHERE id=$1`, it.id)
		}
		time.Sleep(1 * time.Second)
	}
}

func resolveReadyTopic(ch, pr string) string {
	switch ch {
	case "email":
		if pr == "high" { return kafka.TopicReadyEmailHigh } ; return kafka.TopicReadyEmailLow
	case "sms":
		if pr == "high" { return kafka.TopicReadySMSHigh } ; return kafka.TopicReadySMSLow
	case "push":
		if pr == "high" { return kafka.TopicReadyPushHigh } ; return kafka.TopicReadyPushLow
	case "inapp":
		if pr == "high" { return kafka.TopicReadyInAppHigh } ; return kafka.TopicReadyInAppLow
	default:
		return kafka.TopicReadyInAppLow
	}
}

func strconv64(id int64) string { return fmt.Sprintf("%d", id) }