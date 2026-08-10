package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/models"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
)

type Track struct {
	tracks *repository.TrackRepository
	tmpl   *templates.Templates
}

func NewTrack(tracks *repository.TrackRepository, tmpl *templates.Templates) *Track {
	return &Track{tracks: tracks, tmpl: tmpl}
}

func (h *Track) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	tracks, _ := h.tracks.List(r.Context())
	if err := h.tmpl.Render(w, "tracks", map[string]any{
		"User":   user,
		"Tracks": tracks,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Track) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.tmpl.Render(w, "track-form", map[string]any{"User": user}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Track) CreateTrack(w http.ResponseWriter, r *http.Request) {
	t := &models.Track{
		Name:    r.FormValue("name"),
		Country: r.FormValue("country"),
		City:    r.FormValue("city"),
		Indoor:  r.FormValue("indoor") == "on",
	}
	if l, err := strconv.Atoi(r.FormValue("length_meters")); err == nil {
		t.LengthMeters = &l
	}
	if v := r.FormValue("location_url"); v != "" {
		t.LocationURL = &v
	}
	if v := r.FormValue("map_embed"); v != "" {
		t.MapEmbed = &v
	}
	if v := r.FormValue("website"); v != "" {
		t.Website = &v
	}
	if err := h.tracks.Create(r.Context(), t); err != nil {
		http.Error(w, "Failed to create track", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/tracks", http.StatusFound)
}

func (h *Track) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	track, _ := h.tracks.GetByID(r.Context(), id)
	if err := h.tmpl.Render(w, "track-form", map[string]any{"User": user, "Track": track}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Track) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	track, _ := h.tracks.GetByID(r.Context(), id)
	if track == nil {
		http.NotFound(w, r)
		return
	}
	track.Name = r.FormValue("name")
	track.Country = r.FormValue("country")
	track.City = r.FormValue("city")
	track.Indoor = r.FormValue("indoor") == "on"
	if l, err := strconv.Atoi(r.FormValue("length_meters")); err == nil {
		track.LengthMeters = &l
	}
	if v := r.FormValue("location_url"); v != "" {
		track.LocationURL = &v
	} else {
		track.LocationURL = nil
	}
	if v := r.FormValue("map_embed"); v != "" {
		track.MapEmbed = &v
	} else {
		track.MapEmbed = nil
	}
	if v := r.FormValue("website"); v != "" {
		track.Website = &v
	} else {
		track.Website = nil
	}
	if err := h.tracks.Update(r.Context(), track); err != nil {
		http.Error(w, "Failed to update track", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/tracks", http.StatusFound)
}

func (h *Track) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.tracks.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete track", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/tracks", http.StatusFound)
}
