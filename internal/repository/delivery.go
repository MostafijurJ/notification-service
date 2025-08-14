package repository

import (
	"context"

	"github.com/mostafijurj/notification-service/internal/models"
)

// GetDeliveryAttemptsByNotification retrieves all delivery attempts for a notification
func (r *Repository) GetDeliveryAttemptsByNotification(ctx context.Context, notificationID int64) ([]models.DeliveryAttempt, error) {
	var attempts []models.DeliveryAttempt
	if err := r.DB.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("attempt_no DESC").
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// GetDeliveryAttemptsByStatus retrieves delivery attempts by status
func (r *Repository) GetDeliveryAttemptsByStatus(ctx context.Context, status string, limit int) ([]models.DeliveryAttempt, error) {
	var attempts []models.DeliveryAttempt
	if err := r.DB.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at ASC").
		Limit(limit).
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// GetFailedDeliveryAttempts retrieves failed delivery attempts
func (r *Repository) GetFailedDeliveryAttempts(ctx context.Context, limit int) ([]models.DeliveryAttempt, error) {
	var attempts []models.DeliveryAttempt
	if err := r.DB.WithContext(ctx).
		Where("status = ?", "failed").
		Order("created_at ASC").
		Limit(limit).
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// UpdateDeliveryAttemptStatus updates the status of a delivery attempt
func (r *Repository) UpdateDeliveryAttemptStatus(ctx context.Context, id int64, status, errorCode, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorCode != "" {
		updates["error_code"] = errorCode
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	return r.DB.WithContext(ctx).
		Model(&models.DeliveryAttempt{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetDeliveryAttemptStats retrieves delivery attempt statistics
func (r *Repository) GetDeliveryAttemptStats(ctx context.Context, notificationID int64) (map[string]int64, error) {
	var stats struct {
		Total   int64 `gorm:"column:total"`
		Success int64 `gorm:"column:success"`
		Failed  int64 `gorm:"column:failed"`
		Pending int64 `gorm:"column:pending"`
	}

	if err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN status = 'success' THEN 1 END) as success,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
				COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending
			FROM delivery_attempts 
			WHERE notification_id = ?
		`, notificationID).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":   stats.Total,
		"success": stats.Success,
		"failed":  stats.Failed,
		"pending": stats.Pending,
	}, nil
}

// GetRetryableDeliveryAttempts retrieves delivery attempts that can be retried
func (r *Repository) GetRetryableDeliveryAttempts(ctx context.Context, maxAttempts int, limit int) ([]models.DeliveryAttempt, error) {
	var attempts []models.DeliveryAttempt
	if err := r.DB.WithContext(ctx).
		Joins("JOIN notifications ON delivery_attempts.notification_id = notifications.id").
		Where("delivery_attempts.status = ? AND delivery_attempts.attempt_no < ?", "failed", maxAttempts).
		Where("notifications.status = ?", "enqueued").
		Order("delivery_attempts.created_at ASC").
		Limit(limit).
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}
