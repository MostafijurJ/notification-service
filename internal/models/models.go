package models

import (
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	ID             int64          `gorm:"primaryKey;autoIncrement"`
	IdempotencyKey *string        `gorm:"uniqueIndex;type:text"`
	UserID         *int64         `gorm:"index:idx_notifications_user_created,priority:1"`
	CampaignID     *int64         `gorm:"index"`
	TypeKey        string         `gorm:"not null;index"`
	Channel        string         `gorm:"not null;index"`
	Payload        []byte         `gorm:"type:jsonb;not null"`
	Priority       string         `gorm:"not null;index:idx_notifications_status,priority:2"`
	ScheduledAt    *time.Time     `gorm:"index"`
	Status         string         `gorm:"not null;index:idx_notifications_status,priority:1"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index:idx_notifications_user_created,priority:2"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relationships
	Campaign         *Campaign         `gorm:"foreignKey:CampaignID"`
	DeliveryAttempts []DeliveryAttempt `gorm:"foreignKey:NotificationID"`
}

type DeliveryAttempt struct {
	ID                int64     `gorm:"primaryKey;autoIncrement"`
	NotificationID    int64     `gorm:"not null;index"`
	AttemptNo         int       `gorm:"not null"`
	ProviderMessageID *string   `gorm:"type:text"`
	Status            string    `gorm:"not null"`
	ErrorCode         *string   `gorm:"type:text"`
	ErrorMessage      *string   `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`

	// Relationships
	Notification Notification `gorm:"foreignKey:NotificationID"`
}

type UserChannelPreference struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_type_channel,priority:1"`
	TypeKey   string    `gorm:"not null;uniqueIndex:idx_user_type_channel,priority:2"`
	Channel   string    `gorm:"not null;uniqueIndex:idx_user_type_channel,priority:3"`
	OptedIn   bool      `gorm:"not null;default:true"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type UserDNDWindow struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;uniqueIndex"`
	StartTime string    `gorm:"not null;type:time"` // HH:MM:SS
	EndTime   string    `gorm:"not null;type:time"` // HH:MM:SS
	Timezone  string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Group struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`
	Name        string         `gorm:"not null;uniqueIndex"`
	Description *string        `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	Members []GroupMember `gorm:"foreignKey:GroupID"`
}

type GroupMember struct {
	GroupID int64 `gorm:"primaryKey"`
	UserID  int64 `gorm:"primaryKey"`

	// Relationships
	Group Group `gorm:"foreignKey:GroupID"`
}

type Campaign struct {
	ID             int64          `gorm:"primaryKey;autoIncrement"`
	Name           string         `gorm:"not null"`
	TypeKey        string         `gorm:"not null;index"`
	Channel        string         `gorm:"not null;index"`
	SegmentGroupID *int64         `gorm:"index"`
	ScheduledAt    *time.Time     `gorm:"index"`
	Priority       string         `gorm:"not null;default:low;index"`
	Status         string         `gorm:"not null;default:scheduled;index"`
	CreatedBy      *string        `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relationships
	Group         *Group         `gorm:"foreignKey:SegmentGroupID"`
	Notifications []Notification `gorm:"foreignKey:CampaignID"`
}

type NotificationType struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Key      string `gorm:"not null;uniqueIndex"`
	Category string `gorm:"not null"`

	// Relationships
	Templates []Template `gorm:"foreignKey:TypeKey;references:Key"`
}

type Template struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	TypeKey   string    `gorm:"not null;index"`
	Channel   string    `gorm:"not null"`
	Locale    string    `gorm:"not null;default:en"`
	Subject   *string   `gorm:"type:text"`
	Body      string    `gorm:"not null;type:text"`
	Version   int       `gorm:"not null;default:1"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relationships
	Type NotificationType `gorm:"foreignKey:TypeKey;references:Key"`
}

type DeviceToken struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	UserID     int64      `gorm:"not null;index"`
	Platform   string     `gorm:"not null"`
	Token      string     `gorm:"not null;uniqueIndex:idx_provider_token"`
	Provider   string     `gorm:"not null;default:fcm;uniqueIndex:idx_provider_token,priority:1"`
	Enabled    bool       `gorm:"not null;default:true"`
	LastSeenAt *time.Time `gorm:"index"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}

type InAppNotification struct {
	ID        int64          `gorm:"primaryKey;autoIncrement"`
	UserID    int64          `gorm:"not null;index:idx_inapp_user_read_created,priority:1"`
	TypeKey   string         `gorm:"not null;index"`
	Title     *string        `gorm:"type:text"`
	Body      string         `gorm:"not null;type:text"`
	Metadata  []byte         `gorm:"type:jsonb"`
	Read      bool           `gorm:"not null;default:false;index:idx_inapp_user_read_created,priority:2"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_inapp_user_read_created,priority:3"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
