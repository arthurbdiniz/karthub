package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/karthub/karthub/internal/models"
)

type TrackRepository struct {
	db *sql.DB
}

func NewTrackRepository(db *sql.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func (r *TrackRepository) List(ctx context.Context) ([]models.Track, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, country, city, length_meters, indoor, location_url, map_embed, website, created_at, updated_at FROM tracks ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("querying tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tracks []models.Track
	for rows.Next() {
		var t models.Track
		if err := rows.Scan(&t.ID, &t.Name, &t.Country, &t.City, &t.LengthMeters, &t.Indoor, &t.LocationURL, &t.MapEmbed, &t.Website, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning track: %w", err)
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (r *TrackRepository) GetByID(ctx context.Context, id int64) (*models.Track, error) {
	var t models.Track
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, country, city, length_meters, indoor, location_url, map_embed, website, created_at, updated_at FROM tracks WHERE id = ?", id,
	).Scan(&t.ID, &t.Name, &t.Country, &t.City, &t.LengthMeters, &t.Indoor, &t.LocationURL, &t.MapEmbed, &t.Website, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying track: %w", err)
	}
	return &t, nil
}

func (r *TrackRepository) Create(ctx context.Context, t *models.Track) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO tracks (name, country, city, length_meters, indoor, location_url, map_embed, website) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		t.Name, t.Country, t.City, t.LengthMeters, t.Indoor, t.LocationURL, t.MapEmbed, t.Website,
	)
	if err != nil {
		return fmt.Errorf("inserting track: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	t.ID = id
	return nil
}

func (r *TrackRepository) Update(ctx context.Context, t *models.Track) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tracks SET name = ?, country = ?, city = ?, length_meters = ?, indoor = ?, location_url = ?, map_embed = ?, website = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		t.Name, t.Country, t.City, t.LengthMeters, t.Indoor, t.LocationURL, t.MapEmbed, t.Website, t.ID,
	)
	if err != nil {
		return fmt.Errorf("updating track: %w", err)
	}
	return nil
}

func (r *TrackRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tracks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting track: %w", err)
	}
	return nil
}
