package repository

import (
	"context"

	"github.com/mostafijurj/notification-service/internal/models"
)

// UpsertDeviceToken creates or updates a device token
func (r *Repository) UpsertDeviceToken(ctx context.Context, userID int64, platform, token, provider string) error {
	deviceToken := &models.DeviceToken{
		UserID:   userID,
		Platform: platform,
		Token:    token,
		Provider: provider,
	}

	// Use GORM's Save for upsert behavior
	return r.DB.WithContext(ctx).Save(deviceToken).Error
}

// GetDeviceTokensByUserID retrieves all device tokens for a user
func (r *Repository) GetDeviceTokensByUserID(ctx context.Context, userID int64) ([]models.DeviceToken, error) {
	var deviceTokens []models.DeviceToken
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND enabled = ?", userID, true).
		Find(&deviceTokens).Error
	return deviceTokens, err
}

// GetDeviceTokensByPlatform retrieves device tokens for a specific platform
func (r *Repository) GetDeviceTokensByPlatform(ctx context.Context, userID int64, platform string) ([]models.DeviceToken, error) {
	var deviceTokens []models.DeviceToken
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND platform = ? AND enabled = ?", userID, platform, true).
		Find(&deviceTokens).Error
	return deviceTokens, err
}

// DisableDeviceToken disables a device token
func (r *Repository) DisableDeviceToken(ctx context.Context, token string) error {
	return r.DB.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("token = ?", token).
		Update("enabled", false).Error
}

// DeleteDeviceToken deletes a device token
func (r *Repository) DeleteDeviceToken(ctx context.Context, token string) error {
	return r.DB.WithContext(ctx).
		Where("token = ?", token).
		Delete(&models.DeviceToken{}).Error
}

// UpdateDeviceTokenLastSeen updates the last seen timestamp for a device token
func (r *Repository) UpdateDeviceTokenLastSeen(ctx context.Context, token string) error {
	return r.DB.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("token = ?", token).
		Update("last_seen_at", "NOW()").Error
}

// GetDeviceTokenStats retrieves statistics for device tokens
func (r *Repository) GetDeviceTokenStats(ctx context.Context, userID int64) (map[string]int64, error) {
	var stats struct {
		Total   int64 `gorm:"column:total"`
		Enabled int64 `gorm:"column:enabled"`
		FCM     int64 `gorm:"column:fcm"`
		APNS    int64 `gorm:"column:apns"`
	}

	err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN enabled = true THEN 1 END) as enabled,
				COUNT(CASE WHEN provider = 'fcm' THEN 1 END) as fcm,
				COUNT(CASE WHEN provider = 'apns' THEN 1 END) as apns
			FROM device_tokens 
			WHERE user_id = ?
		`, userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":   stats.Total,
		"enabled": stats.Enabled,
		"fcm":     stats.FCM,
		"apns":    stats.APNS,
	}, nil
}
