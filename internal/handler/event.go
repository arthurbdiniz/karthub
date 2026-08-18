package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/models"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Event struct {
	events   *repository.EventRepository
	bookings *repository.BookingRepository
	tracks   *repository.TrackRepository
	drivers  *repository.DriverRepository
	photos   *repository.PhotoRepository
	results  *repository.ResultRepository
	tmpl     *templates.Templates
}

func NewEvent(events *repository.EventRepository, bookings *repository.BookingRepository, tracks *repository.TrackRepository, drivers *repository.DriverRepository, photos *repository.PhotoRepository, results *repository.ResultRepository, tmpl *templates.Templates) *Event {
	return &Event{events: events, bookings: bookings, tracks: tracks, drivers: drivers, photos: photos, results: results, tmpl: tmpl}
}

func (h *Event) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	events, _ := h.events.ListWithTracks(r.Context())
	if err := h.tmpl.Render(w, "events", map[string]any{
		"User":   user,
		"Events": events,
	}); err != nil {
		slog.Error("rendering events", "error", err)
	}
}

func (h *Event) Show(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	event, _ := h.events.GetByID(r.Context(), id)
	if event == nil {
		http.NotFound(w, r)
		return
	}
	bookings, _ := h.bookings.ListByEventWithDrivers(r.Context(), id)
	track, _ := h.tracks.GetByID(r.Context(), event.TrackID)

	// Split bookings by status
	var drivers, waitlisted []repository.BookingWithDriver
	for _, b := range bookings {
		if b.Status == "confirmed" || b.Status == "pending" {
			drivers = append(drivers, b)
		} else {
			waitlisted = append(waitlisted, b)
		}
	}

	var driverID int64
	var myBookingID int64
	hasDriver := false
	canBook := false
	isBooked := false

	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver != nil {
		hasDriver = true
		driverID = driver.ID
		canBook = true
		for _, b := range bookings {
			if b.DriverID == driverID {
				canBook = false
				isBooked = true
				myBookingID = b.ID
				break
			}
		}
	}

	spotsLeft := event.MaxDrivers - len(drivers)
	if spotsLeft < 0 {
		spotsLeft = 0
	}

	// Photos — allowed once lineup is locked (closed, ongoing, completed)
	photos, _ := h.photos.ListByEvent(r.Context(), id)
	canUpload := false
	if driver != nil && event.Status != "open" {
		canUpload, _ = h.photos.DriverParticipated(r.Context(), id, driver.ID)
	}

	// Results — only for completed events
	var results []repository.ResultWithDriver
	if event.Status == "completed" {
		results, _ = h.results.ListByEvent(r.Context(), id)
	}

	if err := h.tmpl.Render(w, "event-detail", map[string]any{
		"User":        user,
		"Event":       event,
		"Track":       track,
		"Drivers":     drivers,
		"Waitlisted":  waitlisted,
		"SpotsLeft":   spotsLeft,
		"HasDriver":   hasDriver,
		"DriverID":    driverID,
		"CanBook":     canBook,
		"IsBooked":    isBooked,
		"MyBookingID": myBookingID,
		"Photos":      photos,
		"CanUpload":   canUpload,
		"Results":     results,
	}); err != nil {
		slog.Error("rendering event detail", "error", err)
	}
}

func (h *Event) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	tracks, _ := h.tracks.List(r.Context())
	if err := h.tmpl.Render(w, "event-form", map[string]any{"User": user, "Tracks": tracks}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Event) CreateEvent(w http.ResponseWriter, r *http.Request) {
	trackID, _ := strconv.ParseInt(r.FormValue("track_id"), 10, 64)
	maxDrivers, _ := strconv.Atoi(r.FormValue("max_drivers"))
	fee, _ := strconv.Atoi(r.FormValue("entry_fee_cents"))
	duration, _ := strconv.Atoi(r.FormValue("duration_minutes"))
	if duration == 0 {
		duration = 60
	}

	e := &models.Event{
		TrackID:         trackID,
		Date:            r.FormValue("date"),
		DurationMinutes: duration,
		MaxDrivers:      maxDrivers,
		EntryFeeCents:   fee,
	}
	if t := r.FormValue("time"); t != "" {
		e.Time = &t
	}
	if n := r.FormValue("notes"); n != "" {
		e.Notes = &n
	}
	if p := r.FormValue("payment_info"); p != "" {
		e.PaymentInfo = &p
	}
	if o := r.FormValue("organizer_contact"); o != "" {
		e.OrganizerContact = &o
	}
	if cid := r.FormValue("championship_id"); cid != "" {
		id, _ := strconv.ParseInt(cid, 10, 64)
		e.ChampionshipID = &id
	}

	if err := h.events.Create(r.Context(), e); err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/events", http.StatusFound)
}

func (h *Event) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	event, _ := h.events.GetByID(r.Context(), id)
	tracks, _ := h.tracks.List(r.Context())
	if err := h.tmpl.Render(w, "event-form", map[string]any{"User": user, "Event": event, "Tracks": tracks}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Event) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	event, _ := h.events.GetByID(r.Context(), id)
	if event == nil {
		http.NotFound(w, r)
		return
	}
	trackID, _ := strconv.ParseInt(r.FormValue("track_id"), 10, 64)
	maxDrivers, _ := strconv.Atoi(r.FormValue("max_drivers"))
	fee, _ := strconv.Atoi(r.FormValue("entry_fee_cents"))
	duration, _ := strconv.Atoi(r.FormValue("duration_minutes"))
	if duration == 0 {
		duration = 60
	}
	event.TrackID = trackID
	event.Date = r.FormValue("date")
	event.DurationMinutes = duration
	event.MaxDrivers = maxDrivers
	event.EntryFeeCents = fee
	if t := r.FormValue("time"); t != "" {
		event.Time = &t
	} else {
		event.Time = nil
	}
	if n := r.FormValue("notes"); n != "" {
		event.Notes = &n
	} else {
		event.Notes = nil
	}
	if p := r.FormValue("payment_info"); p != "" {
		event.PaymentInfo = &p
	} else {
		event.PaymentInfo = nil
	}
	if o := r.FormValue("organizer_contact"); o != "" {
		event.OrganizerContact = &o
	} else {
		event.OrganizerContact = nil
	}
	if err := h.events.Update(r.Context(), event); err != nil {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusFound)
}

func (h *Event) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.events.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/events", http.StatusFound)
}

func (h *Event) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	event, _ := h.events.GetByID(r.Context(), id)
	if event == nil {
		http.NotFound(w, r)
		return
	}

	newStatus := r.FormValue("status")

	// Validate transition
	valid := map[string][]string{
		"open":      {"closed"},
		"closed":    {"open", "ongoing"},
		"ongoing":   {"open", "completed"},
		"completed": {"open"},
	}

	allowed := false
	for _, s := range valid[event.Status] {
		if s == newStatus {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Invalid status transition", http.StatusBadRequest)
		return
	}

	if err := h.events.SetStatus(r.Context(), id, newStatus); err != nil {
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/events/"+strconv.FormatInt(id, 10), http.StatusFound)
}
