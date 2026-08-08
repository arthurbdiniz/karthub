package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/karthub/karthub/internal/models"
)

var (
	ErrEventFull     = errors.New("event is full")
	ErrAlreadyBooked = errors.New("already booked")
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Book(ctx context.Context, eventID, driverID int64) (*models.Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Get max drivers
	var maxDrivers int
	err = tx.QueryRowContext(ctx,
		"SELECT max_drivers FROM events WHERE id = ?", eventID,
	).Scan(&maxDrivers)
	if err != nil {
		return nil, fmt.Errorf("getting event max_drivers: %w", err)
	}

	// Count confirmed + pending (both take a spot)
	var takenCount int
	err = tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM bookings WHERE event_id = ? AND status IN ('confirmed', 'pending')", eventID,
	).Scan(&takenCount)
	if err != nil {
		return nil, fmt.Errorf("counting taken spots: %w", err)
	}

	status := "pending"
	if takenCount >= maxDrivers {
		status = "waitlist"
	}

	// Determine position (next in queue)
	var position int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(position), 0) + 1 FROM bookings WHERE event_id = ? AND status = ?", eventID, status,
	).Scan(&position)
	if err != nil {
		return nil, fmt.Errorf("getting next position: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		"INSERT INTO bookings (event_id, driver_id, status, position) VALUES (?, ?, ?, ?)",
		eventID, driverID, status, position,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	id, _ := result.LastInsertId()
	return &models.Booking{
		ID:       id,
		EventID:  eventID,
		DriverID: driverID,
		Status:   status,
		Position: &position,
	}, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var eventID int64
	err = tx.QueryRowContext(ctx,
		"SELECT event_id FROM bookings WHERE id = ? AND status = 'confirmed'", id,
	).Scan(&eventID)
	if err == sql.ErrNoRows {
		// Just delete waitlist entries directly
		_, err = tx.ExecContext(ctx, "DELETE FROM bookings WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("deleting booking: %w", err)
		}
		return tx.Commit()
	}
	if err != nil {
		return fmt.Errorf("querying booking: %w", err)
	}

	// Delete the cancelled booking
	_, err = tx.ExecContext(ctx, "DELETE FROM bookings WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting booking: %w", err)
	}

	// Promote first waitlisted driver
	_, err = tx.ExecContext(ctx, `
		UPDATE bookings SET status = 'confirmed'
		WHERE id = (
			SELECT id FROM bookings
			WHERE event_id = ? AND status = 'waitlist'
			ORDER BY position ASC
			LIMIT 1
		)
	`, eventID)
	if err != nil {
		return fmt.Errorf("promoting waitlist: %w", err)
	}

	return tx.Commit()
}

func (r *BookingRepository) ListByEvent(ctx context.Context, eventID int64) ([]models.Booking, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, event_id, driver_id, status, position, created_at FROM bookings WHERE event_id = ? ORDER BY status, position", eventID)
	if err != nil {
		return nil, fmt.Errorf("querying bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.EventID, &b.DriverID, &b.Status, &b.Position, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

type BookingWithDriver struct {
	ID             int64
	EventID        int64
	DriverID       int64
	Status         string
	DriverName     string
	DriverNickname *string
	DriverAvatar   *string
}

func (r *BookingRepository) ListByEventWithDrivers(ctx context.Context, eventID int64) ([]BookingWithDriver, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.event_id, b.driver_id, b.status, d.name, d.nickname, d.avatar
		FROM bookings b
		JOIN drivers d ON d.id = b.driver_id
		WHERE b.event_id = ?
		ORDER BY b.status, b.position
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying bookings with drivers: %w", err)
	}
	defer rows.Close()

	var bookings []BookingWithDriver
	for rows.Next() {
		var b BookingWithDriver
		if err := rows.Scan(&b.ID, &b.EventID, &b.DriverID, &b.Status, &b.DriverName, &b.DriverNickname, &b.DriverAvatar); err != nil {
			return nil, fmt.Errorf("scanning booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

func (r *BookingRepository) CountByEvent(ctx context.Context, eventID int64) (confirmed int, waitlisted int, err error) {
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FILTER (WHERE status = 'confirmed'), COUNT(*) FILTER (WHERE status = 'waitlist') FROM bookings WHERE event_id = ?", eventID,
	).Scan(&confirmed, &waitlisted)
	if err != nil {
		return 0, 0, fmt.Errorf("counting bookings: %w", err)
	}
	return confirmed, waitlisted, nil
}

func (r *BookingRepository) Confirm(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE bookings SET status = 'confirmed' WHERE id = ? AND status = 'pending'", id)
	if err != nil {
		return fmt.Errorf("confirming booking: %w", err)
	}
	return nil
}

func (r *BookingRepository) Unconfirm(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE bookings SET status = 'pending' WHERE id = ? AND status = 'confirmed'", id)
	if err != nil {
		return fmt.Errorf("unconfirming booking: %w", err)
	}
	return nil
}

func (r *BookingRepository) MoveToWaitlist(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE bookings SET status = 'waitlist' WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("moving to waitlist: %w", err)
	}
	return nil
}

func (r *BookingRepository) Remove(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM bookings WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("removing booking: %w", err)
	}
	return nil
}

func (r *BookingRepository) Promote(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE bookings SET status = 'pending' WHERE id = ? AND status = 'waitlist'", id)
	if err != nil {
		return fmt.Errorf("promoting booking: %w", err)
	}
	return nil
}
