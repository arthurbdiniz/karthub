package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/karthub/karthub/internal/models"
)

type DriverRepository struct {
	db *sql.DB
}

func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) List(ctx context.Context) ([]models.Driver, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, name, nickname, avatar, country_code, bio, joined_at, created_at, updated_at FROM drivers ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("querying drivers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var drivers []models.Driver
	for rows.Next() {
		var d models.Driver
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.Nickname, &d.Avatar, &d.CountryCode, &d.Bio, &d.JoinedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning driver: %w", err)
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}

func (r *DriverRepository) GetByID(ctx context.Context, id int64) (*models.Driver, error) {
	var d models.Driver
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, name, nickname, avatar, country_code, bio, joined_at, created_at, updated_at FROM drivers WHERE id = ?", id,
	).Scan(&d.ID, &d.UserID, &d.Name, &d.Nickname, &d.Avatar, &d.CountryCode, &d.Bio, &d.JoinedAt, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying driver: %w", err)
	}
	return &d, nil
}

func (r *DriverRepository) Create(ctx context.Context, d *models.Driver) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO drivers (user_id, name, nickname, avatar, country_code, bio, joined_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		d.UserID, d.Name, d.Nickname, d.Avatar, d.CountryCode, d.Bio, d.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting driver: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	d.ID = id
	return nil
}

func (r *DriverRepository) Update(ctx context.Context, d *models.Driver) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE drivers SET name = ?, nickname = ?, avatar = ?, country_code = ?, bio = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		d.Name, d.Nickname, d.Avatar, d.CountryCode, d.Bio, d.ID,
	)
	if err != nil {
		return fmt.Errorf("updating driver: %w", err)
	}
	return nil
}

func (r *DriverRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM drivers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting driver: %w", err)
	}
	return nil
}

func (r *DriverRepository) GetByUserID(ctx context.Context, userID int64) (*models.Driver, error) {
	var d models.Driver
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, name, nickname, avatar, country_code, bio, joined_at, created_at, updated_at FROM drivers WHERE user_id = ?", userID,
	).Scan(&d.ID, &d.UserID, &d.Name, &d.Nickname, &d.Avatar, &d.CountryCode, &d.Bio, &d.JoinedAt, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying driver by user: %w", err)
	}
	return &d, nil
}
