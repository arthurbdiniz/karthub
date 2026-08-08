package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/karthub/karthub/internal/models"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) List(ctx context.Context) ([]models.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, championship_id, track_id, date, time, duration_minutes, max_drivers, entry_fee_cents, payment_info, organizer_contact, status, notes, created_at, updated_at FROM events ORDER BY date DESC")
	if err != nil {
		return nil, fmt.Errorf("querying events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.PaymentInfo, &e.OrganizerContact, &e.Status, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	var e models.Event
	err := r.db.QueryRowContext(ctx,
		"SELECT id, championship_id, track_id, date, time, duration_minutes, max_drivers, entry_fee_cents, payment_info, organizer_contact, status, notes, created_at, updated_at FROM events WHERE id = ?", id,
	).Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.PaymentInfo, &e.OrganizerContact, &e.Status, &e.Notes, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying event: %w", err)
	}
	return &e, nil
}

func (r *EventRepository) Upcoming(ctx context.Context, limit int) ([]models.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, championship_id, track_id, date, time, duration_minutes, max_drivers, entry_fee_cents, payment_info, organizer_contact, status, notes, created_at, updated_at FROM events WHERE date >= date('now') ORDER BY date ASC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("querying upcoming events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.PaymentInfo, &e.OrganizerContact, &e.Status, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) Create(ctx context.Context, e *models.Event) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO events (championship_id, track_id, date, time, duration_minutes, max_drivers, entry_fee_cents, payment_info, organizer_contact, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		e.ChampionshipID, e.TrackID, e.Date, e.Time, e.DurationMinutes, e.MaxDrivers, e.EntryFeeCents, e.PaymentInfo, e.OrganizerContact, e.Notes,
	)
	if err != nil {
		return fmt.Errorf("inserting event: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	e.ID = id
	return nil
}

func (r *EventRepository) Update(ctx context.Context, e *models.Event) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE events SET championship_id = ?, track_id = ?, date = ?, time = ?, duration_minutes = ?, max_drivers = ?, entry_fee_cents = ?, payment_info = ?, organizer_contact = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		e.ChampionshipID, e.TrackID, e.Date, e.Time, e.DurationMinutes, e.MaxDrivers, e.EntryFeeCents, e.PaymentInfo, e.OrganizerContact, e.Notes, e.ID,
	)
	if err != nil {
		return fmt.Errorf("updating event: %w", err)
	}
	return nil
}

func (r *EventRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM events WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting event: %w", err)
	}
	return nil
}

func (r *EventRepository) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE events SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("updating event status: %w", err)
	}
	return nil
}

func (r *EventRepository) GetByBookingID(ctx context.Context, bookingID int64) (*models.Event, error) {
	var e models.Event
	err := r.db.QueryRowContext(ctx, `
		SELECT e.id, e.championship_id, e.track_id, e.date, e.time, e.duration_minutes, e.max_drivers, e.entry_fee_cents, e.status, e.notes, e.created_at, e.updated_at
		FROM events e
		JOIN bookings b ON b.event_id = e.id
		WHERE b.id = ?
	`, bookingID).Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.PaymentInfo, &e.OrganizerContact, &e.Status, &e.Notes, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying event by booking: %w", err)
	}
	return &e, nil
}

type EventWithTrack struct {
	ID              int64
	ChampionshipID  *int64
	TrackID         int64
	Date            string
	Time            *string
	DurationMinutes int
	MaxDrivers      int
	EntryFeeCents   int
	Status          string
	Notes           *string
	TrackName       string
	TrackCity       string
	BookingCount    int
	WaitlistCount   int
}

func (r *EventRepository) ListWithTracks(ctx context.Context) ([]EventWithTrack, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.championship_id, e.track_id, e.date, e.time, e.duration_minutes, e.max_drivers, e.entry_fee_cents, e.status, e.notes,
			t.name, t.city,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'confirmed') as booking_count,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'waitlist') as waitlist_count
		FROM events e
		JOIN tracks t ON t.id = e.track_id
		ORDER BY e.date DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying events with tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []EventWithTrack
	for rows.Next() {
		var e EventWithTrack
		if err := rows.Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.Status, &e.Notes, &e.TrackName, &e.TrackCity, &e.BookingCount, &e.WaitlistCount); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) UpcomingWithTracks(ctx context.Context, limit int) ([]EventWithTrack, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.championship_id, e.track_id, e.date, e.time, e.duration_minutes, e.max_drivers, e.entry_fee_cents, e.status, e.notes,
			t.name, t.city,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'confirmed') as booking_count,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'waitlist') as waitlist_count
		FROM events e
		JOIN tracks t ON t.id = e.track_id
		WHERE e.date >= date('now')
		ORDER BY e.date ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("querying upcoming events with tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []EventWithTrack
	for rows.Next() {
		var e EventWithTrack
		if err := rows.Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.Status, &e.Notes, &e.TrackName, &e.TrackCity, &e.BookingCount, &e.WaitlistCount); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) PastWithTracks(ctx context.Context, limit int) ([]EventWithTrack, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.championship_id, e.track_id, e.date, e.time, e.duration_minutes, e.max_drivers, e.entry_fee_cents, e.status, e.notes,
			t.name, t.city,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'confirmed') as booking_count,
			(SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'waitlist') as waitlist_count
		FROM events e
		JOIN tracks t ON t.id = e.track_id
		WHERE e.date < date('now')
		ORDER BY e.date DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("querying past events with tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []EventWithTrack
	for rows.Next() {
		var e EventWithTrack
		if err := rows.Scan(&e.ID, &e.ChampionshipID, &e.TrackID, &e.Date, &e.Time, &e.DurationMinutes, &e.MaxDrivers, &e.EntryFeeCents, &e.Status, &e.Notes, &e.TrackName, &e.TrackCity, &e.BookingCount, &e.WaitlistCount); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
