package repository

import "context"

type GroupCreate struct {
	Name        string
	Description *string
}

func (r *Repository) CreateGroup(ctx context.Context, g GroupCreate) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `INSERT INTO groups (name, description) VALUES ($1,$2) RETURNING id`, g.Name, g.Description).Scan(&id)
	return id, err
}

func (r *Repository) AddGroupMember(ctx context.Context, groupID int64, userID int64) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO group_members (group_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, groupID, userID)
	return err
}

func (r *Repository) ListGroupMembers(ctx context.Context, groupID int64, limit int, offset int) ([]int64, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT user_id FROM group_members WHERE group_id=$1 ORDER BY user_id LIMIT $2 OFFSET $3`, groupID, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var res []int64
	for rows.Next() { var uid int64; if err := rows.Scan(&uid); err != nil { return nil, err }; res = append(res, uid) }
	return res, nil
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
	var id int64
	err := r.DB.QueryRowContext(ctx, `
	INSERT INTO campaigns (name, type_key, channel, segment_group_id, scheduled_at, priority, status, created_by)
	VALUES ($1,$2,$3,$4,CAST(NULLIF($5,'') AS TIMESTAMPTZ),$6,'scheduled',$7)
	RETURNING id
	`, c.Name, c.TypeKey, c.Channel, c.SegmentGroupID, c.ScheduledAt, c.Priority, c.CreatedBy).Scan(&id)
	return id, err
}