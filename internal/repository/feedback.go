package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type FeedbackRepository struct {
	db *sql.DB
}

func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

type FeedbackWithDriver struct {
	ID             int64
	DriverID       int64
	EventID        *int64
	Message        string
	CreatedAt      string
	DriverNickname *string
	DriverName     string
	DriverAvatar   *string
}

func (r *FeedbackRepository) Create(ctx context.Context, driverID int64, eventID *int64, message string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO feedback (driver_id, event_id, message) VALUES (?, ?, ?)",
		driverID, eventID, message,
	)
	if err != nil {
		return fmt.Errorf("inserting feedback: %w", err)
	}
	return nil
}

func (r *FeedbackRepository) List(ctx context.Context) ([]FeedbackWithDriver, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.id, f.driver_id, f.event_id, f.message, f.created_at, d.nickname, d.name, d.avatar
		FROM feedback f
		JOIN drivers d ON d.id = f.driver_id
		ORDER BY f.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying feedback: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []FeedbackWithDriver
	for rows.Next() {
		var fb FeedbackWithDriver
		if err := rows.Scan(&fb.ID, &fb.DriverID, &fb.EventID, &fb.Message, &fb.CreatedAt, &fb.DriverNickname, &fb.DriverName, &fb.DriverAvatar); err != nil {
			return nil, fmt.Errorf("scanning feedback: %w", err)
		}
		items = append(items, fb)
	}
	return items, rows.Err()
}
