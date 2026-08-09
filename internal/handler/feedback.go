package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Feedback struct {
	feedback *repository.FeedbackRepository
	drivers  *repository.DriverRepository
	tmpl     *templates.Templates
}

func NewFeedback(feedback *repository.FeedbackRepository, drivers *repository.DriverRepository, tmpl *templates.Templates) *Feedback {
	return &Feedback{feedback: feedback, drivers: drivers, tmpl: tmpl}
}

func (h *Feedback) Form(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.tmpl.Render(w, "feedback", map[string]any{"User": user}); err != nil {
		slog.Error("rendering feedback form", "error", err)
	}
}

func (h *Feedback) Submit(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver == nil {
		http.Error(w, "No driver profile", http.StatusForbidden)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		if err := h.tmpl.Render(w, "feedback", map[string]any{"User": user, "Error": "Message is required"}); err != nil {
			slog.Error("rendering feedback form", "error", err)
		}
		return
	}

	var eventID *int64
	if v := r.FormValue("event_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		eventID = &id
	}

	if err := h.feedback.Create(r.Context(), driver.ID, eventID, message); err != nil {
		http.Error(w, "Failed to submit feedback", http.StatusInternalServerError)
		return
	}

	if err := h.tmpl.Render(w, "feedback", map[string]any{"User": user, "Success": true}); err != nil {
		slog.Error("rendering feedback form", "error", err)
	}
}

func (h *Feedback) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	items, _ := h.feedback.List(r.Context())
	if err := h.tmpl.Render(w, "feedback-list", map[string]any{"User": user, "Feedback": items}); err != nil {
		slog.Error("rendering feedback list", "error", err)
	}
}
