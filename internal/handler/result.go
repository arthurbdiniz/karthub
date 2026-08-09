package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/repository"
)

type Result struct {
	results *repository.ResultRepository
}

func NewResult(results *repository.ResultRepository) *Result {
	return &Result{results: results}
}

func (h *Result) Save(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil || (user.Role != "admin" && user.Role != "organizer") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

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
	w.Header().Set("HX-Trigger", "refreshResults")
	_, _ = w.Write([]byte(`<span class="text-green-600 text-xs">✓ Saved</span>`))
}

func (h *Result) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil || (user.Role != "admin" && user.Role != "organizer") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	eventID, _ := strconv.ParseInt(chi.URLParam(r, "eventID"), 10, 64)
	driverID, _ := strconv.ParseInt(r.FormValue("driver_id"), 10, 64)

	if err := h.results.Delete(r.Context(), eventID, driverID); err != nil {
		http.Error(w, "Failed to delete result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", "refreshResults")
	_, _ = w.Write([]byte(`<span class="text-green-600 text-xs">✓ Reset</span>`))
}

func (h *Result) Table(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.ParseInt(chi.URLParam(r, "eventID"), 10, 64)
	results, _ := h.results.ListByEvent(r.Context(), eventID)

	w.Header().Set("Content-Type", "text/html")
	if len(results) == 0 {
		_, _ = w.Write([]byte(`<p class="text-gray-500 text-sm">No results entered yet.</p>`))
		return
	}

	html := `<table class="w-full text-sm"><thead><tr class="border-b text-left text-gray-500"><th class="py-2 pr-3">Pos</th><th class="py-2 pr-3">Driver</th><th class="py-2 pr-3">Best Lap</th><th class="py-2 pr-3" title="Fastest single lap in the race">Fastest</th><th class="py-2 pr-3" title="Did Not Finish — didn't complete the race">DNF</th><th class="py-2">Penalty</th></tr></thead><tbody class="divide-y divide-gray-100">`

	for _, res := range results {
		pos := strconv.Itoa(res.Position)
		medal := pos
		switch res.Position {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		}

		name := res.DriverName
		if res.DriverNickname != nil {
			name = *res.DriverNickname
		}
		if res.DriverCountry != nil && *res.DriverCountry != "" {
			name += " " + countryFlag(*res.DriverCountry)
		}

		lap := "-"
		if res.BestLapTime != nil {
			lap = *res.BestLapTime
		}
		fastest := ""
		if res.FastestLap {
			fastest = `<span title="Fastest lap of the race">⚡</span>`
		}
		dnf := ""
		if res.DNF {
			dnf = `<span title="Did Not Finish">❌</span>`
		}
		penalty := ""
		if res.PenaltySeconds > 0 {
			penalty = "+" + strconv.Itoa(res.PenaltySeconds) + "s"
		}

		html += `<tr><td class="py-2 pr-3 font-bold">` + medal + `</td><td class="py-2 pr-3"><span class="font-medium">` + name + `</span></td><td class="py-2 pr-3">` + lap + `</td><td class="py-2 pr-3">` + fastest + `</td><td class="py-2 pr-3">` + dnf + `</td><td class="py-2">` + penalty + `</td></tr>`
	}

	html += `</tbody></table>`
	_, _ = w.Write([]byte(html))
}

func countryFlag(code string) string {
	if len(code) != 2 {
		return ""
	}
	r0 := rune(code[0]) - 'A' + 0x1F1E6
	r1 := rune(code[1]) - 'A' + 0x1F1E6
	return string([]rune{r0, r1})
}
