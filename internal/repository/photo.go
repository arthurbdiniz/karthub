package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/karthub/karthub/internal/models"
)

type PhotoRepository struct {
	db *sql.DB
}

func NewPhotoRepository(db *sql.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Create(ctx context.Context, p *models.EventPhoto) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO event_photos (event_id, driver_id, filename, original_name) VALUES (?, ?, ?, ?)",
		p.EventID, p.DriverID, p.Filename, p.OriginalName,
	)
	if err != nil {
		return fmt.Errorf("inserting photo: %w", err)
	}
	id, _ := result.LastInsertId()
	p.ID = id
	return nil
}

func (r *PhotoRepository) ListByEvent(ctx context.Context, eventID int64) ([]models.EventPhoto, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, event_id, driver_id, filename, original_name, created_at FROM event_photos WHERE event_id = ? ORDER BY created_at DESC", eventID)
	if err != nil {
		return nil, fmt.Errorf("querying photos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var photos []models.EventPhoto
	for rows.Next() {
		var p models.EventPhoto
		if err := rows.Scan(&p.ID, &p.EventID, &p.DriverID, &p.Filename, &p.OriginalName, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning photo: %w", err)
		}
		photos = append(photos, p)
	}
	return photos, rows.Err()
}

func (r *PhotoRepository) Delete(ctx context.Context, id int64) (string, error) {
	var filename string
	err := r.db.QueryRowContext(ctx, "SELECT filename FROM event_photos WHERE id = ?", id).Scan(&filename)
	if err != nil {
		return "", fmt.Errorf("querying photo: %w", err)
	}
	_, err = r.db.ExecContext(ctx, "DELETE FROM event_photos WHERE id = ?", id)
	if err != nil {
		return "", fmt.Errorf("deleting photo: %w", err)
	}
	return filename, nil
}

// DriverParticipated checks if a driver has a confirmed booking for this event.
func (r *PhotoRepository) DriverParticipated(ctx context.Context, eventID, driverID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM bookings WHERE event_id = ? AND driver_id = ?", eventID, driverID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
