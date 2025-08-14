package repository

import (
	"context"
	"time"

	"github.com/mostafijurj/notification-service/internal/models"
)

type GroupCreate struct {
	Name        string
	Description *string
}

func (r *Repository) CreateGroup(ctx context.Context, g GroupCreate) (int64, error) {
	group := &models.Group{
		Name:        g.Name,
		Description: g.Description,
	}

	if err := r.DB.WithContext(ctx).Create(group).Error; err != nil {
		return 0, err
	}

	return group.ID, nil
}

func (r *Repository) AddGroupMember(ctx context.Context, groupID int64, userID int64) error {
	groupMember := &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}

	return r.DB.WithContext(ctx).Create(groupMember).Error
}

func (r *Repository) ListGroupMembers(ctx context.Context, groupID int64, limit int, offset int) ([]int64, error) {
	var groupMembers []models.GroupMember
	if err := r.DB.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("user_id").
		Limit(limit).
		Offset(offset).
		Find(&groupMembers).Error; err != nil {
		return nil, err
	}

	var userIDs []int64
	for _, member := range groupMembers {
		userIDs = append(userIDs, member.UserID)
	}

	return userIDs, nil
}

type CampaignCreate struct {
	Name           string
	TypeKey        string
	Channel        string
	SegmentGroupID *int64
	ScheduledAt    *string
	Priority       string
	CreatedBy      *string
}

func (r *Repository) CreateCampaign(ctx context.Context, c CampaignCreate) (int64, error) {
	campaign := &models.Campaign{
		Name:           c.Name,
		TypeKey:        c.TypeKey,
		Channel:        c.Channel,
		SegmentGroupID: c.SegmentGroupID,
		Priority:       c.Priority,
		Status:         "scheduled",
		CreatedBy:      c.CreatedBy,
	}

	// Parse scheduled time if provided
	if c.ScheduledAt != nil && *c.ScheduledAt != "" {
		if scheduledTime, err := time.Parse(time.RFC3339, *c.ScheduledAt); err == nil {
			campaign.ScheduledAt = &scheduledTime
		}
	}

	if err := r.DB.WithContext(ctx).Create(campaign).Error; err != nil {
		return 0, err
	}

	return campaign.ID, nil
}

// GetGroupByID retrieves a group by ID
func (r *Repository) GetGroupByID(ctx context.Context, groupID int64) (*models.Group, error) {
	var group models.Group
	if err := r.DB.WithContext(ctx).
		Preload("Members").
		First(&group, groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupsByUserID retrieves all groups a user belongs to
func (r *Repository) GetGroupsByUserID(ctx context.Context, userID int64) ([]models.Group, error) {
	var groups []models.Group
	if err := r.DB.WithContext(ctx).
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// GetUsersInGroup retrieves all users in a specific group
func (r *Repository) GetUsersInGroup(ctx context.Context, groupID int64) ([]int64, error) {
	var userIDs []int64
	if err := r.DB.WithContext(ctx).
		Model(&models.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

// RemoveGroupMember removes a user from a group
func (r *Repository) RemoveGroupMember(ctx context.Context, groupID, userID int64) error {
	return r.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).Error
}

// GetCampaignByID retrieves a campaign by ID
func (r *Repository) GetCampaignByID(ctx context.Context, campaignID int64) (*models.Campaign, error) {
	var campaign models.Campaign
	if err := r.DB.WithContext(ctx).
		Preload("Group").
		Preload("Notifications").
		First(&campaign, campaignID).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

// GetCampaignsByStatus retrieves campaigns by status
func (r *Repository) GetCampaignsByStatus(ctx context.Context, status string) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	if err := r.DB.WithContext(ctx).
		Where("status = ?", status).
		Preload("Group").
		Order("created_at DESC").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetCampaignsByType retrieves campaigns by notification type
func (r *Repository) GetCampaignsByType(ctx context.Context, typeKey string) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	if err := r.DB.WithContext(ctx).
		Where("type_key = ?", typeKey).
		Preload("Group").
		Order("created_at DESC").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetCampaignsByChannel retrieves campaigns by channel
func (r *Repository) GetCampaignsByChannel(ctx context.Context, channel string) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	if err := r.DB.WithContext(ctx).
		Where("channel = ?", channel).
		Preload("Group").
		Order("created_at DESC").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetScheduledCampaigns retrieves campaigns that are scheduled to run
func (r *Repository) GetScheduledCampaigns(ctx context.Context) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	if err := r.DB.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", "scheduled", "NOW()").
		Preload("Group").
		Order("scheduled_at ASC").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// UpdateCampaignStatus updates the status of a campaign
func (r *Repository) UpdateCampaignStatus(ctx context.Context, campaignID int64, status string) error {
	return r.DB.WithContext(ctx).
		Model(&models.Campaign{}).
		Where("id = ?", campaignID).
		Update("status", status).Error
}
