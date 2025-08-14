package repository

import (
	"context"

	"github.com/mostafijurj/notification-service/internal/models"
)

// GetTemplate retrieves a template by type, channel, and locale
func (r *Repository) GetTemplate(ctx context.Context, typeKey, channel, locale string) (*models.Template, error) {
	var template models.Template
	err := r.DB.WithContext(ctx).
		Where("type_key = ? AND channel = ? AND locale = ? AND is_active = ?", typeKey, channel, locale, true).
		Order("version DESC").
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetTemplatesByType retrieves all templates for a notification type
func (r *Repository) GetTemplatesByType(ctx context.Context, typeKey string) ([]models.Template, error) {
	var templates []models.Template
	if err := r.DB.WithContext(ctx).
		Where("type_key = ? AND is_active = ?", typeKey, true).
		Order("channel, locale, version DESC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetTemplatesByChannel retrieves all templates for a specific channel
func (r *Repository) GetTemplatesByChannel(ctx context.Context, channel string) ([]models.Template, error) {
	var templates []models.Template
	if err := r.DB.WithContext(ctx).
		Where("channel = ? AND is_active = ?", channel, true).
		Order("type_key, locale, version DESC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// CreateTemplate creates a new template
func (r *Repository) CreateTemplate(ctx context.Context, template *models.Template) error {
	return r.DB.WithContext(ctx).Create(template).Error
}

// UpdateTemplate updates an existing template
func (r *Repository) UpdateTemplate(ctx context.Context, template *models.Template) error {
	return r.DB.WithContext(ctx).Save(template).Error
}

// DeactivateTemplate deactivates a template
func (r *Repository) DeactivateTemplate(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).
		Model(&models.Template{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// GetTemplateVersions retrieves all versions of a template
func (r *Repository) GetTemplateVersions(ctx context.Context, typeKey, channel, locale string) ([]models.Template, error) {
	var templates []models.Template
	if err := r.DB.WithContext(ctx).
		Where("type_key = ? AND channel = ? AND locale = ?", typeKey, channel, locale).
		Order("version DESC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetNotificationTypes retrieves all notification types
func (r *Repository) GetNotificationTypes(ctx context.Context) ([]models.NotificationType, error) {
	var types []models.NotificationType
	if err := r.DB.WithContext(ctx).
		Preload("Templates").
		Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// GetNotificationTypeByKey retrieves a notification type by key
func (r *Repository) GetNotificationTypeByKey(ctx context.Context, key string) (*models.NotificationType, error) {
	var notificationType models.NotificationType
	err := r.DB.WithContext(ctx).
		Where("key = ?", key).
		Preload("Templates").
		First(&notificationType).Error
	if err != nil {
		return nil, err
	}
	return &notificationType, nil
}
