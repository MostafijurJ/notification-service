package models

import "time"

type Notification struct {
	ID             int64
	IdempotencyKey *string
	UserID         *int64
	CampaignID     *int64
	TypeKey        string
	Channel        string
	Payload        []byte
	Priority       string
	ScheduledAt    *time.Time
	Status         string
	CreatedAt      time.Time
}

type DeliveryAttempt struct {
	ID                int64
	NotificationID    int64
	AttemptNo         int
	ProviderMessageID *string
	Status            string
	ErrorCode         *string
	ErrorMessage      *string
	CreatedAt         time.Time
}

type UserChannelPreference struct {
	UserID   int64
	TypeKey  string
	Channel  string
	OptedIn  bool
	UpdatedAt time.Time
}

type UserDNDWindow struct {
	UserID    int64
	StartTime string // HH:MM:SS
	EndTime   string // HH:MM:SS
	Timezone  string
}

type Group struct {
	ID          int64
	Name        string
	Description *string
}

type Campaign struct {
	ID             int64
	Name           string
	TypeKey        string
	Channel        string
	SegmentGroupID *int64
	ScheduledAt    *time.Time
	Priority       string
	Status         string
	CreatedBy      *string
	CreatedAt      time.Time
}