package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/karthub/karthub/internal/config"
	"github.com/karthub/karthub/internal/models"
)

const cookieName = "karthub_session"

type Manager struct {
	db  *sql.DB
	cfg config.SessionConfig
}

func NewManager(db *sql.DB, cfg config.SessionConfig) *Manager {
	return &Manager{db: db, cfg: cfg}
}

func (m *Manager) Create(w http.ResponseWriter, userID int64) error {
	id, err := generateID()
	if err != nil {
		return fmt.Errorf("generating session id: %w", err)
	}

	expiresAt := time.Now().UTC().Add(time.Duration(m.cfg.MaxAge) * time.Second).Format("2006-01-02 15:04:05")

	_, err = m.db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		id, userID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   m.cfg.MaxAge,
		HttpOnly: m.cfg.HTTPOnly,
		Secure:   m.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (m *Manager) GetUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, nil
	}

	var user models.User
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	err = m.db.QueryRow(`
		SELECT u.id, u.email, u.name, u.role, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ? AND s.expires_at > ?
	`, cookie.Value, now).Scan(
		&user.ID, &user.Email, &user.Name,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying session: %w", err)
	}

	return &user, nil
}

func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}

	_, err = m.db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	return nil
}

func generateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
