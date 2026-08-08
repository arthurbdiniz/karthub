package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Dashboard struct {
	events   *repository.EventRepository
	bookings *repository.BookingRepository
	drivers  *repository.DriverRepository
	results  *repository.ResultRepository
	users    *repository.UserRepository
	tmpl     *templates.Templates
}

func NewDashboard(events *repository.EventRepository, bookings *repository.BookingRepository, drivers *repository.DriverRepository, results *repository.ResultRepository, users *repository.UserRepository, tmpl *templates.Templates) *Dashboard {
	return &Dashboard{events: events, bookings: bookings, drivers: drivers, results: results, users: users, tmpl: tmpl}
}

func (h *Dashboard) Index(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	events, _ := h.events.UpcomingWithTracks(r.Context(), 5)
	pastEvents, _ := h.events.PastWithTracks(r.Context(), 5)
	myDriver, _ := h.drivers.GetByUserID(r.Context(), user.ID)

	var myStats *repository.DriverStats
	if myDriver != nil {
		myStats, _ = h.results.StatsByDriver(r.Context(), myDriver.ID)
	}

	if err := h.tmpl.Render(w, "dashboard", map[string]any{
		"User":       user,
		"Events":     events,
		"PastEvents": pastEvents,
		"MyDriver":   myDriver,
		"MyStats":    myStats,
	}); err != nil {
		slog.Error("rendering dashboard", "error", err)
	}
}

func (h *Dashboard) Admin(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.tmpl.Render(w, "admin", map[string]any{
		"User": user,
	}); err != nil {
		slog.Error("rendering admin", "error", err)
	}
}

func (h *Dashboard) Users(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	users, _ := h.users.List(r.Context())
	if err := h.tmpl.Render(w, "users", map[string]any{
		"User":  user,
		"Users": users,
	}); err != nil {
		slog.Error("rendering users", "error", err)
	}
}

func (h *Dashboard) SetRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	role := r.FormValue("role")

	if role != "user" && role != "organizer" && role != "admin" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	if err := h.users.SetRole(r.Context(), id, role); err != nil {
		http.Error(w, "Failed to set role", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusFound)
}
