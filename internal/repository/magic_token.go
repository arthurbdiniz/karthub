package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const tokenExpiry = 15 * time.Minute

var ErrTokenInvalid = errors.New("token is invalid or expired")

type MagicTokenRepository struct {
	db *sql.DB
}

func NewMagicTokenRepository(db *sql.DB) *MagicTokenRepository {
	return &MagicTokenRepository{db: db}
}

func (r *MagicTokenRepository) Create(ctx context.Context, email string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(tokenExpiry).Format("2006-01-02 15:04:05")

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO magic_tokens (token, email, expires_at) VALUES (?, ?, ?)",
		token, email, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("inserting magic token: %w", err)
	}

	return token, nil
}

func (r *MagicTokenRepository) Consume(ctx context.Context, token string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	var email string
	err = tx.QueryRowContext(ctx,
		"SELECT email FROM magic_tokens WHERE token = ? AND used = 0 AND expires_at > ?",
		token, now,
	).Scan(&email)
	if err == sql.ErrNoRows {
		return "", ErrTokenInvalid
	}
	if err != nil {
		return "", fmt.Errorf("querying token: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE magic_tokens SET used = 1 WHERE token = ?", token,
	)
	if err != nil {
		return "", fmt.Errorf("marking token used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("committing transaction: %w", err)
	}

	return email, nil
}

// Cleanup removes expired tokens. Call periodically.
func (r *MagicTokenRepository) Cleanup(ctx context.Context) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM magic_tokens WHERE expires_at < ? OR used = 1", now,
	)
	return err
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
