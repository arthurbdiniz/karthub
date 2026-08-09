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
	// Get next sort_order for this event
	var maxOrder int
	_ = r.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(sort_order), 0) FROM event_photos WHERE event_id = ?", p.EventID).Scan(&maxOrder)

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO event_photos (event_id, driver_id, filename, original_name, sort_order) VALUES (?, ?, ?, ?, ?)",
		p.EventID, p.DriverID, p.Filename, p.OriginalName, maxOrder+1,
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
		"SELECT id, event_id, driver_id, filename, original_name, created_at FROM event_photos WHERE event_id = ? ORDER BY sort_order ASC, created_at DESC", eventID)
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

func (r *PhotoRepository) MoveUp(ctx context.Context, eventID, photoID int64) error {
	// Get current photo's sort_order
	var currentOrder int
	err := r.db.QueryRowContext(ctx, "SELECT sort_order FROM event_photos WHERE id = ?", photoID).Scan(&currentOrder)
	if err != nil {
		return err
	}

	// Find the photo above it (lower sort_order)
	var prevID int64
	var prevOrder int
	err = r.db.QueryRowContext(ctx,
		"SELECT id, sort_order FROM event_photos WHERE event_id = ? AND sort_order < ? ORDER BY sort_order DESC LIMIT 1",
		eventID, currentOrder).Scan(&prevID, &prevOrder)
	if err != nil {
		return nil // already at top
	}

	// Swap their sort_orders
	_, _ = r.db.ExecContext(ctx, "UPDATE event_photos SET sort_order = ? WHERE id = ?", prevOrder, photoID)
	_, _ = r.db.ExecContext(ctx, "UPDATE event_photos SET sort_order = ? WHERE id = ?", currentOrder, prevID)
	return nil
}

func (r *PhotoRepository) MoveDown(ctx context.Context, eventID, photoID int64) error {
	var currentOrder int
	err := r.db.QueryRowContext(ctx, "SELECT sort_order FROM event_photos WHERE id = ?", photoID).Scan(&currentOrder)
	if err != nil {
		return err
	}

	var nextID int64
	var nextOrder int
	err = r.db.QueryRowContext(ctx,
		"SELECT id, sort_order FROM event_photos WHERE event_id = ? AND sort_order > ? ORDER BY sort_order ASC LIMIT 1",
		eventID, currentOrder).Scan(&nextID, &nextOrder)
	if err != nil {
		return nil // already at bottom
	}

	_, _ = r.db.ExecContext(ctx, "UPDATE event_photos SET sort_order = ? WHERE id = ?", nextOrder, photoID)
	_, _ = r.db.ExecContext(ctx, "UPDATE event_photos SET sort_order = ? WHERE id = ?", currentOrder, nextID)
	return nil
}

func (r *PhotoRepository) Reorder(ctx context.Context, ids []int64) error {
	for i, id := range ids {
		_, err := r.db.ExecContext(ctx, "UPDATE event_photos SET sort_order = ? WHERE id = ?", i+1, id)
		if err != nil {
			return fmt.Errorf("updating sort_order: %w", err)
		}
	}
	return nil
}

func (r *PhotoRepository) GetEventIDByFilename(ctx context.Context, filename string) (int64, error) {
	var eventID int64
	err := r.db.QueryRowContext(ctx,
		"SELECT event_id FROM event_photos WHERE filename = ?", filename,
	).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("querying event by filename: %w", err)
	}
	return eventID, nil
}
