package repository

import (
	"context"
	"time"

	"github.com/mostafijurj/notification-service/internal/models"
)

type NotificationCreate struct {
	IdempotencyKey *string
	UserID         *int64
	CampaignID     *int64
	TypeKey        string
	Channel        string
	PayloadJSON    []byte
	Priority       string
	ScheduledAt    *string
	Status         string
}

func (r *Repository) CreateNotification(ctx context.Context, n NotificationCreate) (int64, error) {
	notification := &models.Notification{
		IdempotencyKey: n.IdempotencyKey,
		UserID:         n.UserID,
		CampaignID:     n.CampaignID,
		TypeKey:        n.TypeKey,
		Channel:        n.Channel,
		Payload:        n.PayloadJSON,
		Priority:       n.Priority,
		Status:         n.Status,
	}

	// Parse scheduled time if provided
	if n.ScheduledAt != nil && *n.ScheduledAt != "" {
		if scheduledTime, err := time.Parse(time.RFC3339, *n.ScheduledAt); err == nil {
			notification.ScheduledAt = &scheduledTime
		}
	}

	if err := r.DB.WithContext(ctx).Create(notification).Error; err != nil {
		return 0, err
	}

	return notification.ID, nil
}

func (r *Repository) UpdateNotificationStatus(ctx context.Context, id int64, status string) error {
	return r.DB.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) InsertDeliveryAttempt(ctx context.Context, notificationID int64, attemptNo int, providerMsgID *string, status string, errCode *string, errMsg *string) error {
	deliveryAttempt := &models.DeliveryAttempt{
		NotificationID:    notificationID,
		AttemptNo:         attemptNo,
		ProviderMessageID: providerMsgID,
		Status:            status,
		ErrorCode:         errCode,
		ErrorMessage:      errMsg,
	}

	return r.DB.WithContext(ctx).Create(deliveryAttempt).Error
}

func (r *Repository) CreateInApp(ctx context.Context, userID int64, typeKey, title, body string, metaJSON []byte) (int64, error) {
	inAppNotification := &models.InAppNotification{
		UserID:   userID,
		TypeKey:  typeKey,
		Body:     body,
		Metadata: metaJSON,
	}

	if title != "" {
		inAppNotification.Title = &title
	}

	if err := r.DB.WithContext(ctx).Create(inAppNotification).Error; err != nil {
		return 0, err
	}

	return inAppNotification.ID, nil
}

type InApp struct {
	ID        int64
	UserID    int64
	TypeKey   string
	Title     *string
	Body      string
	Read      bool
	CreatedAt time.Time
}

func (r *Repository) ListInApp(ctx context.Context, userID int64, onlyUnread bool, limit int) ([]InApp, error) {
	var notifications []models.InAppNotification
	query := r.DB.WithContext(ctx).Where("user_id = ?", userID)

	if onlyUnread {
		query = query.Where("read = ?", false)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	var result []InApp
	for _, n := range notifications {
		result = append(result, InApp{
			ID:        n.ID,
			UserID:    n.UserID,
			TypeKey:   n.TypeKey,
			Title:     n.Title,
			Body:      n.Body,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		})
	}

	return result, nil
}

// GetInAppNotificationByID retrieves a specific in-app notification
func (r *Repository) GetInAppNotificationByID(ctx context.Context, id int64) (*models.InAppNotification, error) {
	var notification models.InAppNotification
	err := r.DB.WithContext(ctx).
		Where("id = ?", id).
		First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// DeleteInAppNotification deletes an in-app notification
func (r *Repository) DeleteInAppNotification(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.InAppNotification{}).Error
}

// GetInAppNotificationStats retrieves statistics for in-app notifications
func (r *Repository) GetInAppNotificationStats(ctx context.Context, userID int64) (map[string]int64, error) {
	var stats struct {
		Total  int64 `gorm:"column:total"`
		Read   int64 `gorm:"column:read"`
		Unread int64 `gorm:"column:unread"`
	}

	err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN read = true THEN 1 END) as read,
				COUNT(CASE WHEN read = false THEN 1 END) as unread
			FROM inapp_notifications 
			WHERE user_id = ?
		`, userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":  stats.Total,
		"read":   stats.Read,
		"unread": stats.Unread,
	}, nil
}

func (r *Repository) MarkInAppRead(ctx context.Context, id int64, read bool) error {
	return r.DB.WithContext(ctx).Model(&models.InAppNotification{}).
		Where("id = ?", id).
		Update("read", read).Error
}

// GetNotificationsByStatus retrieves notifications by status with pagination
func (r *Repository) GetNotificationsByStatus(ctx context.Context, status string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.DB.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// GetScheduledNotifications retrieves notifications that are due for delivery
func (r *Repository) GetScheduledNotifications(ctx context.Context, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.DB.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", "scheduled", time.Now()).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// GetNotificationByID retrieves a notification by ID with delivery attempts
func (r *Repository) GetNotificationByID(ctx context.Context, id int64) (*models.Notification, error) {
	var notification models.Notification
	err := r.DB.WithContext(ctx).
		Preload("DeliveryAttempts").
		Preload("Campaign").
		First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetNotificationsByUserID retrieves notifications for a specific user
func (r *Repository) GetNotificationsByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// GetNotificationsByCampaignID retrieves notifications for a specific campaign
func (r *Repository) GetNotificationsByCampaignID(ctx context.Context, campaignID int64) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.DB.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Order("created_at ASC").
		Find(&notifications).Error
	return notifications, err
}

// GetNotificationStats retrieves notification statistics
func (r *Repository) GetNotificationStats(ctx context.Context, userID int64) (map[string]int64, error) {
	var stats struct {
		Total     int64 `gorm:"column:total"`
		Sent      int64 `gorm:"column:sent"`
		Failed    int64 `gorm:"column:failed"`
		Pending   int64 `gorm:"column:pending"`
		Scheduled int64 `gorm:"column:scheduled"`
	}

	err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
				COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
				COUNT(CASE WHEN status = 'scheduled' THEN 1 END) as scheduled
			FROM notifications 
			WHERE user_id = ?
		`, userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":     stats.Total,
		"sent":      stats.Sent,
		"failed":    stats.Failed,
		"pending":   stats.Pending,
		"scheduled": stats.Scheduled,
	}, nil
}

// CreateBulkNotifications creates multiple notifications at once
func (r *Repository) CreateBulkNotifications(ctx context.Context, notifications []models.Notification) error {
	return r.DB.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// GetNotificationsByBatch retrieves notifications by batch ID
func (r *Repository) GetNotificationsByBatch(ctx context.Context, batchID string) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.DB.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("created_at ASC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// UpdateBulkNotificationStatus updates status for multiple notifications
func (r *Repository) UpdateBulkNotificationStatus(ctx context.Context, ids []int64, status string) error {
	return r.DB.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// GetBulkNotificationStats retrieves statistics for bulk notifications
func (r *Repository) GetBulkNotificationStats(ctx context.Context, batchID string) (map[string]int64, error) {
	var stats struct {
		Total     int64 `gorm:"column:total"`
		Sent      int64 `gorm:"column:sent"`
		Failed    int64 `gorm:"column:failed"`
		Pending   int64 `gorm:"column:pending"`
		Scheduled int64 `gorm:"column:scheduled"`
	}

	if err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
				COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
				COUNT(CASE WHEN status = 'scheduled' THEN 1 END) as scheduled
			FROM notifications 
			WHERE batch_id = ?
		`, batchID).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":     stats.Total,
		"sent":      stats.Sent,
		"failed":    stats.Failed,
		"pending":   stats.Pending,
		"scheduled": stats.Scheduled,
	}, nil
}
