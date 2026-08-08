package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/repository"
)

type Result struct {
	results *repository.ResultRepository
}

func NewResult(results *repository.ResultRepository) *Result {
	return &Result{results: results}
}

func (h *Result) Save(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.ParseInt(chi.URLParam(r, "eventID"), 10, 64)
	driverID, _ := strconv.ParseInt(r.FormValue("driver_id"), 10, 64)
	position, _ := strconv.Atoi(r.FormValue("position"))
	penaltySeconds, _ := strconv.Atoi(r.FormValue("penalty_seconds"))
	fastestLap := r.FormValue("fastest_lap") == "on"
	dnf := r.FormValue("dnf") == "on"

	var bestLapTime *string
	if v := r.FormValue("best_lap_time"); v != "" {
		bestLapTime = &v
	}
	var notes *string
	if v := r.FormValue("notes"); v != "" {
		notes = &v
	}

	err := h.results.Upsert(r.Context(), eventID, driverID, position, bestLapTime, fastestLap, dnf, penaltySeconds, notes)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`<span class="text-red-600 text-xs">Error</span>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<span class="text-green-600 text-xs">✓ Saved</span>`))
}
