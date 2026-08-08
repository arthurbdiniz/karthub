package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/karthub/karthub/internal/models"
	"github.com/karthub/karthub/internal/session"
)

type contextKey string

const userContextKey contextKey = "user"

type Auth struct {
	sessions *session.Manager
	db       *sql.DB
}

func NewAuth(sessions *session.Manager, db *sql.DB) *Auth {
	return &Auth{sessions: sessions, db: db}
}

func (a *Auth) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.sessions.GetUser(r)
		if err != nil || user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProfileRequired redirects to /setup if the user has no driver profile.
// Skips the check if already on /setup.
func (a *Auth) ProfileRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/setup") || strings.HasPrefix(r.URL.Path, "/logout") {
			next.ServeHTTP(w, r)
			return
		}

		user := UserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		var count int
		a.db.QueryRow("SELECT COUNT(*) FROM drivers WHERE user_id = ?", user.ID).Scan(&count)
		if count == 0 {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Auth) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminOrOrganizer allows admin and organizer roles.
func (a *Auth) AdminOrOrganizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || (user.Role != "admin" && user.Role != "organizer") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userContextKey).(*models.User)
	return user
}
