package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/push"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Push struct {
	repo    *repository.PushRepository
	service *push.Service
	tmpl    *templates.Templates
}

func NewPush(repo *repository.PushRepository, service *push.Service, tmpl *templates.Templates) *Push {
	return &Push{repo: repo, service: service, tmpl: tmpl}
}

// Subscribe saves the push subscription from the browser.
func (h *Push) Subscribe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	endpoint := r.FormValue("endpoint")
	p256dh := r.FormValue("p256dh")
	auth := r.FormValue("auth")

	if endpoint == "" || p256dh == "" || auth == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	if err := h.repo.Save(r.Context(), user.ID, endpoint, p256dh, auth); err != nil {
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// VAPIDKey returns the public VAPID key for the browser.
func (h *Push) VAPIDKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(h.service.VAPIDPublicKey()))
}

// Send triggers a push notification to all subscribers (admin only).
func (h *Push) Send(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil || (user.Role != "admin" && user.Role != "organizer") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	url := r.FormValue("url")
	audience := r.FormValue("audience")
	eventID := r.FormValue("event_id")

	if title == "" || body == "" {
		http.Error(w, "Title and body required", http.StatusBadRequest)
		return
	}

	payload := push.Payload{Title: title, Body: body, URL: url}

	switch audience {
	case "event":
		if eventID != "" {
			id, _ := strconv.ParseInt(eventID, 10, 64)
			go h.service.SendToEvent(context.Background(), id, payload)
		}
	default:
		go h.service.SendAll(context.Background(), payload)
	}

	if err := h.tmpl.Render(w, "notify", map[string]any{"User": user, "Success": true}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// Page renders the notification form.
func (h *Push) Page(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.tmpl.Render(w, "notify", map[string]any{"User": user}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
