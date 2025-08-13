package repository

import (
	"context"
)

type PreferenceUpsert struct {
	UserID  int64
	TypeKey string
	Channel string
	OptedIn bool
}

func (r *Repository) UpsertPreference(ctx context.Context, p PreferenceUpsert) error {
	_, err := r.DB.ExecContext(ctx, `
	INSERT INTO user_channel_preferences (user_id, type_key, channel, opted_in)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id, type_key, channel)
	DO UPDATE SET opted_in = EXCLUDED.opted_in, updated_at = NOW()
	`, p.UserID, p.TypeKey, p.Channel, p.OptedIn)
	return err
}

func (r *Repository) GetPreference(ctx context.Context, userID int64, typeKey, channel string) (bool, error) {
	var optedIn bool
	err := r.DB.QueryRowContext(ctx, `
	SELECT opted_in FROM user_channel_preferences WHERE user_id=$1 AND type_key=$2 AND channel=$3
	`, userID, typeKey, channel).Scan(&optedIn)
	if err != nil {
		// default to opted in if not found
		return true, nil
	}
	return optedIn, nil
}

func (r *Repository) UpsertDND(ctx context.Context, userID int64, start, end, tz string) error {
	_, err := r.DB.ExecContext(ctx, `
	INSERT INTO user_dnd_windows (user_id, start_time, end_time, timezone)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id)
	DO UPDATE SET start_time=EXCLUDED.start_time, end_time=EXCLUDED.end_time, timezone=EXCLUDED.timezone
	`, userID, start, end, tz)
	return err
}

func (r *Repository) GetDND(ctx context.Context, userID int64) (start, end, tz string, found bool, err error) {
	err = r.DB.QueryRowContext(ctx, `
	SELECT start_time::text, end_time::text, timezone FROM user_dnd_windows WHERE user_id=$1
	`, userID).Scan(&start, &end, &tz)
	if err != nil {
		return "", "", "", false, nil
	}
	return start, end, tz, true, nil
}