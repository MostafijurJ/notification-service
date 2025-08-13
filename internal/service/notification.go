package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mostafijurj/notification-service/internal/cache"
	"github.com/mostafijurj/notification-service/internal/kafka"
	"github.com/mostafijurj/notification-service/internal/repository"
)

type NotificationRequest struct {
	UserID      int64             `json:"user_id"`
	TypeKey     string            `json:"type_key"`
	Channels    []string          `json:"channels"`
	Payload     map[string]any    `json:"payload"`
	Priority    string            `json:"priority"`      // high|low
	ScheduledAt *string           `json:"scheduled_at"`  // RFC3339
}

// NotificationRepo defines the subset of repository used by NotificationService
// This enables mocking in unit tests.
type NotificationRepo interface {
	CreateNotification(ctx context.Context, n repository.NotificationCreate) (int64, error)
	GetDND(ctx context.Context, userID int64) (start, end, tz string, found bool, err error)
}

type NotificationService struct {
	Repo      NotificationRepo
	Redis     *cache.Redis
	Broker    string
	Produce   func(broker, topic, key, value string) error
}

func NewNotificationService(repo NotificationRepo, redis *cache.Redis, broker string) *NotificationService {
	return &NotificationService{Repo: repo, Redis: redis, Broker: broker, Produce: kafka.ProduceMessage}
}

func (s *NotificationService) Enqueue(ctx context.Context, req NotificationRequest, idempotencyKey *string) ([]int64, error) {
	if req.Priority == "" { req.Priority = "low" }
	ids := make([]int64, 0, len(req.Channels))
	payloadBytes, _ := json.Marshal(req.Payload)

	// DND enforcement and preference per channel
	start, end, tz, hasDND, _ := s.Repo.GetDND(ctx, req.UserID)
	now := time.Now()
	var scheduleAt *string
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		scheduleAt = req.ScheduledAt
	} else if hasDND && !isHighPriorityAllowed(req.TypeKey) {
		if inDND(now, start, end, tz) {
			// schedule right after DND end today
			d := nextDndEndTime(now, end, tz)
			iso := d.Format(time.RFC3339)
			scheduleAt = &iso
		}
	}

	for _, ch := range req.Channels {
		id, err := s.Repo.CreateNotification(ctx, repository.NotificationCreate{
			IdempotencyKey: idempotencyKey,
			UserID:         &req.UserID,
			TypeKey:        req.TypeKey,
			Channel:        ch,
			PayloadJSON:    payloadBytes,
			Priority:       req.Priority,
			ScheduledAt:    scheduleAt,
			Status:         statusFromSchedule(scheduleAt),
		})
		if err != nil { return nil, err }
		ids = append(ids, id)

		if scheduleAt == nil { // immediate routing
			readyTopic := resolveReadyTopic(ch, req.Priority)
			msgKey := fmt.Sprintf("user:%d", req.UserID)
			msgVal := fmt.Sprintf("%d", id)
			_ = s.Produce(s.Broker, readyTopic, msgKey, msgVal)
		}
	}
	return ids, nil
}

func statusFromSchedule(s *string) string { if s == nil { return "enqueued" } ; return "scheduled" }

func resolveReadyTopic(channel, priority string) string {
	switch channel {
	case "email":
		if priority == "high" { return kafka.TopicReadyEmailHigh }; return kafka.TopicReadyEmailLow
	case "sms":
		if priority == "high" { return kafka.TopicReadySMSHigh }; return kafka.TopicReadySMSLow
	case "push":
		if priority == "high" { return kafka.TopicReadyPushHigh }; return kafka.TopicReadyPushLow
	case "inapp":
		if priority == "high" { return kafka.TopicReadyInAppHigh }; return kafka.TopicReadyInAppLow
	default:
		return kafka.TopicReadyInAppLow
	}
}

// Simplified: allowlist of types that bypass DND
func isHighPriorityAllowed(typeKey string) bool {
	switch typeKey {
	case "auth.otp", "security.alert":
		return true
	default:
		return false
	}
}

func inDND(now time.Time, start string, end string, tz string) bool {
	loc, err := time.LoadLocation(tz); if err != nil { loc = time.UTC }
	n := now.In(loc)
	st, _ := time.ParseInLocation("15:04:05", start, loc)
	ed, _ := time.ParseInLocation("15:04:05", end, loc)
	startToday := time.Date(n.Year(), n.Month(), n.Day(), st.Hour(), st.Minute(), st.Second(), 0, loc)
	endToday := time.Date(n.Year(), n.Month(), n.Day(), ed.Hour(), ed.Minute(), ed.Second(), 0, loc)
	if endToday.Before(startToday) { // window crosses midnight
		if n.After(startToday) { return true }
		endToday = endToday.Add(24 * time.Hour)
		startToday = startToday.Add(-24 * time.Hour)
	}
	return !n.Before(startToday) && !n.After(endToday)
}

func nextDndEndTime(now time.Time, end string, tz string) time.Time {
	loc, err := time.LoadLocation(tz); if err != nil { loc = time.UTC }
	n := now.In(loc)
	ed, _ := time.ParseInLocation("15:04:05", end, loc)
	endToday := time.Date(n.Year(), n.Month(), n.Day(), ed.Hour(), ed.Minute(), ed.Second(), 0, loc)
	if endToday.Before(n) { endToday = endToday.Add(24 * time.Hour) }
	return endToday
}