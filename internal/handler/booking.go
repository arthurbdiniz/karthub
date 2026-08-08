package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Booking struct {
	bookings *repository.BookingRepository
	events   *repository.EventRepository
	tmpl     *templates.Templates
}

func NewBooking(bookings *repository.BookingRepository, events *repository.EventRepository, tmpl *templates.Templates) *Booking {
	return &Booking{bookings: bookings, events: events, tmpl: tmpl}
}

func (h *Booking) Create(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.ParseInt(r.FormValue("event_id"), 10, 64)
	driverID, _ := strconv.ParseInt(r.FormValue("driver_id"), 10, 64)

	event, _ := h.events.GetByID(r.Context(), eventID)
	if event == nil || event.Status != "open" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div class="text-red-600 text-sm">Bookings are not open for this event</div>`))
		return
	}

	_, err := h.bookings.Book(r.Context(), eventID, driverID)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div class="text-red-600 text-sm">` + err.Error() + `</div>`))
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) Cancel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	// Check if event is still open
	event, _ := h.events.GetByBookingID(r.Context(), id)
	if event != nil && event.Status != "open" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div class="text-red-600 text-sm">Cannot cancel — event is no longer open</div>`))
		return
	}

	if err := h.bookings.Cancel(r.Context(), id); err != nil {
		http.Error(w, "Failed to cancel", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) Confirm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.bookings.Confirm(r.Context(), id); err != nil {
		http.Error(w, "Failed to confirm", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) Unconfirm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.bookings.Unconfirm(r.Context(), id); err != nil {
		http.Error(w, "Failed to unconfirm", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) MoveToWaitlist(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.bookings.MoveToWaitlist(r.Context(), id); err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) Remove(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.bookings.Remove(r.Context(), id); err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Booking) Promote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.bookings.Promote(r.Context(), id); err != nil {
		http.Error(w, "Failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
