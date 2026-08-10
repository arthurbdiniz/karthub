package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type PushSubscription struct {
	ID       int64
	UserID   int64
	Endpoint string
	P256dh   string
	Auth     string
}

type PushRepository struct {
	db *sql.DB
}

func NewPushRepository(db *sql.DB) *PushRepository {
	return &PushRepository{db: db}
}

func (r *PushRepository) Save(ctx context.Context, userID int64, endpoint, p256dh, auth string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, key_p256dh, key_auth)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(endpoint) DO UPDATE SET key_p256dh = excluded.key_p256dh, key_auth = excluded.key_auth
	`, userID, endpoint, p256dh, auth)
	if err != nil {
		return fmt.Errorf("saving push subscription: %w", err)
	}
	return nil
}

func (r *PushRepository) ListAll(ctx context.Context) ([]PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, endpoint, key_p256dh, key_auth FROM push_subscriptions")
	if err != nil {
		return nil, fmt.Errorf("querying subscriptions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var subs []PushSubscription
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth); err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func (r *PushRepository) Delete(ctx context.Context, endpoint string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM push_subscriptions WHERE endpoint = ?", endpoint)
	return err
}

func (r *PushRepository) ListByUserIDs(ctx context.Context, userIDs []int64) ([]PushSubscription, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	// Build placeholders
	placeholders := ""
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, endpoint, key_p256dh, key_auth FROM push_subscriptions WHERE user_id IN ("+placeholders+")", args...)
	if err != nil {
		return nil, fmt.Errorf("querying subscriptions by users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var subs []PushSubscription
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth); err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func (r *PushRepository) ListByEventID(ctx context.Context, eventID int64) ([]PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ps.id, ps.user_id, ps.endpoint, ps.key_p256dh, ps.key_auth
		FROM push_subscriptions ps
		JOIN drivers d ON d.user_id = ps.user_id
		JOIN bookings b ON b.driver_id = d.id
		WHERE b.event_id = ?
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying subscriptions by event: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var subs []PushSubscription
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth); err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}
