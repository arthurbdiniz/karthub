package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/karthub/karthub/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) CreateFromEmail(ctx context.Context, email string) (*models.User, error) {
	name := strings.Split(email, "@")[0]

	// First user gets admin role
	role := "user"
	var count int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		role = "admin"
	}

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO users (email, name, role) VALUES (?, ?, ?)",
		email, name, role,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	id, _ := result.LastInsertId()
	return r.GetByID(ctx, id)
}

func (r *UserRepository) SetRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", role, id)
	if err != nil {
		return fmt.Errorf("setting user role: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, email, name, role, created_at, updated_at FROM users ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
