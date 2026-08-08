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
	h.tmpl.Render(w, "tracks", map[string]any{
		"User":   user,
		"Tracks": tracks,
	})
}

func (h *Track) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	h.tmpl.Render(w, "track-form", map[string]any{"User": user})
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
	h.tracks.Create(r.Context(), t)
	http.Redirect(w, r, "/tracks", http.StatusFound)
}

func (h *Track) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	track, _ := h.tracks.GetByID(r.Context(), id)
	h.tmpl.Render(w, "track-form", map[string]any{"User": user, "Track": track})
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
	h.tracks.Update(r.Context(), track)
	http.Redirect(w, r, "/tracks", http.StatusFound)
}

func (h *Track) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.tracks.Delete(r.Context(), id)
	http.Redirect(w, r, "/admin/tracks", http.StatusFound)
}
