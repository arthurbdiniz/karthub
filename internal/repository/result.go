package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type ResultRepository struct {
	db *sql.DB
}

func NewResultRepository(db *sql.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

type ResultWithDriver struct {
	ID             int64
	EventID        int64
	DriverID       int64
	Position       int
	BestLapTime    *string
	FastestLap     bool
	DNF            bool
	PenaltySeconds int
	Notes          *string
	DriverName     string
	DriverNickname *string
	DriverAvatar   *string
}

func (r *ResultRepository) ListByEvent(ctx context.Context, eventID int64) ([]ResultWithDriver, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.event_id, r.driver_id, r.position, r.best_lap_time, r.fastest_lap, r.dnf, r.penalty_seconds, r.notes,
			d.name, d.nickname, d.avatar
		FROM results r
		JOIN drivers d ON d.id = r.driver_id
		WHERE r.event_id = ?
		ORDER BY r.position ASC
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("querying results: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []ResultWithDriver
	for rows.Next() {
		var res ResultWithDriver
		if err := rows.Scan(&res.ID, &res.EventID, &res.DriverID, &res.Position, &res.BestLapTime, &res.FastestLap, &res.DNF, &res.PenaltySeconds, &res.Notes, &res.DriverName, &res.DriverNickname, &res.DriverAvatar); err != nil {
			return nil, fmt.Errorf("scanning result: %w", err)
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

func (r *ResultRepository) Upsert(ctx context.Context, eventID, driverID int64, position int, bestLapTime *string, fastestLap, dnf bool, penaltySeconds int, notes *string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO results (event_id, driver_id, position, best_lap_time, fastest_lap, dnf, penalty_seconds, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(event_id, driver_id) DO UPDATE SET
			position = excluded.position,
			best_lap_time = excluded.best_lap_time,
			fastest_lap = excluded.fastest_lap,
			dnf = excluded.dnf,
			penalty_seconds = excluded.penalty_seconds,
			notes = excluded.notes
	`, eventID, driverID, position, bestLapTime, fastestLap, dnf, penaltySeconds, notes)
	if err != nil {
		return fmt.Errorf("upserting result: %w", err)
	}
	return nil
}

type DriverRaceHistory struct {
	EventID      int64
	EventDate    string
	TrackName    string
	Position     int
	BestLapTime  *string
	FastestLap   bool
	DNF          bool
	TotalDrivers int
}

func (r *ResultRepository) HistoryByDriver(ctx context.Context, driverID int64) ([]DriverRaceHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.event_id, e.date, t.name, r.position, r.best_lap_time, r.fastest_lap, r.dnf,
			(SELECT COUNT(*) FROM results r2 WHERE r2.event_id = r.event_id) as total_drivers
		FROM results r
		JOIN events e ON e.id = r.event_id
		JOIN tracks t ON t.id = e.track_id
		WHERE r.driver_id = ?
		ORDER BY e.date DESC
	`, driverID)
	if err != nil {
		return nil, fmt.Errorf("querying driver history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []DriverRaceHistory
	for rows.Next() {
		var h DriverRaceHistory
		if err := rows.Scan(&h.EventID, &h.EventDate, &h.TrackName, &h.Position, &h.BestLapTime, &h.FastestLap, &h.DNF, &h.TotalDrivers); err != nil {
			return nil, fmt.Errorf("scanning history: %w", err)
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

type DriverStats struct {
	TotalRaces  int
	Wins        int
	Podiums     int
	FastestLaps int
	DNFs        int
}

func (r *ResultRepository) StatsByDriver(ctx context.Context, driverID int64) (*DriverStats, error) {
	var s DriverStats
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			SUM(CASE WHEN position = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN position <= 3 THEN 1 ELSE 0 END),
			SUM(CASE WHEN fastest_lap = 1 THEN 1 ELSE 0 END),
			SUM(CASE WHEN dnf = 1 THEN 1 ELSE 0 END)
		FROM results WHERE driver_id = ?
	`, driverID).Scan(&s.TotalRaces, &s.Wins, &s.Podiums, &s.FastestLaps, &s.DNFs)
	if err != nil {
		return nil, fmt.Errorf("querying driver stats: %w", err)
	}
	return &s, nil
}
