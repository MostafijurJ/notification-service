package repository

import (
	"context"

	"github.com/mostafijurj/notification-service/internal/models"
)

type PreferenceUpsert struct {
	UserID  int64
	TypeKey string
	Channel string
	OptedIn bool
}

func (r *Repository) UpsertPreference(ctx context.Context, p PreferenceUpsert) error {
	preference := &models.UserChannelPreference{
		UserID:  p.UserID,
		TypeKey: p.TypeKey,
		Channel: p.Channel,
		OptedIn: p.OptedIn,
	}

	// Use GORM's Upsert (Create or Update)
	return r.DB.WithContext(ctx).Save(preference).Error
}

func (r *Repository) GetPreference(ctx context.Context, userID int64, typeKey, channel string) (bool, error) {
	var preference models.UserChannelPreference
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND type_key = ? AND channel = ?", userID, typeKey, channel).
		First(&preference).Error

	if err != nil {
		// default to opted in if not found
		return true, nil
	}
	return preference.OptedIn, nil
}

func (r *Repository) UpsertDND(ctx context.Context, userID int64, start, end, tz string) error {
	dndWindow := &models.UserDNDWindow{
		UserID:    userID,
		StartTime: start,
		EndTime:   end,
		Timezone:  tz,
	}

	// Use GORM's Upsert (Create or Update)
	return r.DB.WithContext(ctx).Save(dndWindow).Error
}

func (r *Repository) GetDND(ctx context.Context, userID int64) (start, end, tz string, found bool, err error) {
	var dndWindow models.UserDNDWindow
	err = r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&dndWindow).Error

	if err != nil {
		return "", "", "", false, nil
	}
	return dndWindow.StartTime, dndWindow.EndTime, dndWindow.Timezone, true, nil
}

// GetUserPreferences retrieves all preferences for a user
func (r *Repository) GetUserPreferences(ctx context.Context, userID int64) ([]models.UserChannelPreference, error) {
	var preferences []models.UserChannelPreference
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&preferences).Error
	return preferences, err
}

// GetUserDNDWindow retrieves DND window for a user
func (r *Repository) GetUserDNDWindow(ctx context.Context, userID int64) (*models.UserDNDWindow, error) {
	var dndWindow models.UserDNDWindow
	err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&dndWindow).Error

	if err != nil {
		return nil, err
	}
	return &dndWindow, nil
}

// GetUserPreferencesByType retrieves preferences for a specific notification type
func (r *Repository) GetUserPreferencesByType(ctx context.Context, userID int64, typeKey string) ([]models.UserChannelPreference, error) {
	var preferences []models.UserChannelPreference
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND type_key = ?", userID, typeKey).
		Find(&preferences).Error
	return preferences, err
}

// GetUserPreferencesByChannel retrieves preferences for a specific channel
func (r *Repository) GetUserPreferencesByChannel(ctx context.Context, userID int64, channel string) ([]models.UserChannelPreference, error) {
	var preferences []models.UserChannelPreference
	if err := r.DB.WithContext(ctx).
		Where("user_id = ? AND channel = ?", userID, channel).
		Find(&preferences).Error; err != nil {
		return nil, err
	}
	return preferences, nil
}

// DeleteUserPreference deletes a specific user preference
func (r *Repository) DeleteUserPreference(ctx context.Context, userID int64, typeKey, channel string) error {
	return r.DB.WithContext(ctx).
		Where("user_id = ? AND type_key = ? AND channel = ?", userID, typeKey, channel).
		Delete(&models.UserChannelPreference{}).Error
}

// DeleteUserDND deletes DND settings for a user
func (r *Repository) DeleteUserDND(ctx context.Context, userID int64) error {
	return r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.UserDNDWindow{}).Error
}
