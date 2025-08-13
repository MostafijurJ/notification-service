package repository

import (
	"context"
	"database/sql"
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
	var id int64
	err := r.DB.QueryRowContext(ctx, `
	INSERT INTO notifications (idempotency_key, user_id, campaign_id, type_key, channel, payload, priority, scheduled_at, status)
	VALUES ($1,$2,$3,$4,$5,$6,$7,CAST(NULLIF($8,'') AS TIMESTAMPTZ),$9)
	RETURNING id
	`, n.IdempotencyKey, n.UserID, n.CampaignID, n.TypeKey, n.Channel, n.PayloadJSON, n.Priority, n.ScheduledAt, n.Status).Scan(&id)
	return id, err
}

func (r *Repository) UpdateNotificationStatus(ctx context.Context, id int64, status string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *Repository) InsertDeliveryAttempt(ctx context.Context, notificationID int64, attemptNo int, providerMsgID *string, status string, errCode *string, errMsg *string) error {
	_, err := r.DB.ExecContext(ctx, `
	INSERT INTO delivery_attempts (notification_id, attempt_no, provider_message_id, status, error_code, error_message)
	VALUES ($1,$2,$3,$4,$5,$6)
	`, notificationID, attemptNo, providerMsgID, status, errCode, errMsg)
	return err
}

func (r *Repository) CreateInApp(ctx context.Context, userID int64, typeKey, title, body string, metaJSON []byte) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `
	INSERT INTO inapp_notifications (user_id, type_key, title, body, metadata)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id
	`, userID, typeKey, title, body, metaJSON).Scan(&id)
	return id, err
}

type InApp struct {
	ID        int64
	UserID    int64
	TypeKey   string
	Title     sql.NullString
	Body      string
	Read      bool
	CreatedAt string
}

func (r *Repository) ListInApp(ctx context.Context, userID int64, onlyUnread bool, limit int) ([]InApp, error) {
	query := `SELECT id, user_id, type_key, title, body, read, created_at::text FROM inapp_notifications WHERE user_id=$1`
	if onlyUnread {
		query += " AND read=false"
	}
	query += " ORDER BY created_at DESC LIMIT $2"
	rows, err := r.DB.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []InApp
	for rows.Next() {
		var it InApp
		if err := rows.Scan(&it.ID, &it.UserID, &it.TypeKey, &it.Title, &it.Body, &it.Read, &it.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, it)
	}
	return res, nil
}

func (r *Repository) MarkInAppRead(ctx context.Context, id int64, read bool) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE inapp_notifications SET read=$2 WHERE id=$1`, id, read)
	return err
}