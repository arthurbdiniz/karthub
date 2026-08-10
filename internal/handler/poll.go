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

type Poll struct {
	polls   *repository.PollRepository
	drivers *repository.DriverRepository
	tmpl    *templates.Templates
}

func NewPoll(polls *repository.PollRepository, drivers *repository.DriverRepository, tmpl *templates.Templates) *Poll {
	return &Poll{polls: polls, drivers: drivers, tmpl: tmpl}
}

func (h *Poll) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	polls, _ := h.polls.ListWithCounts(r.Context())
	if err := h.tmpl.Render(w, "polls", map[string]any{"User": user, "Polls": polls}); err != nil {
		slog.Error("rendering polls", "error", err)
	}
}

func (h *Poll) Show(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())

	poll, _ := h.polls.GetByID(r.Context(), id)
	if poll == nil {
		http.NotFound(w, r)
		return
	}

	options, _ := h.polls.GetOptions(r.Context(), id)
	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)

	var myVotes []int64
	if driver != nil {
		myVotes, _ = h.polls.GetUserVotes(r.Context(), id, driver.ID)
	}

	// Total votes and max for progress bar
	totalVotes := 0
	maxVotes := 0
	for _, o := range options {
		totalVotes += o.VoteCount
		if o.VoteCount > maxVotes {
			maxVotes = o.VoteCount
		}
	}

	if err := h.tmpl.Render(w, "poll-detail", map[string]any{
		"User":       user,
		"Poll":       poll,
		"Options":    options,
		"MyVotes":    myVotes,
		"TotalVotes": totalVotes,
		"MaxVotes":   maxVotes,
		"HasDriver":  driver != nil,
	}); err != nil {
		slog.Error("rendering poll", "error", err)
	}
}

func (h *Poll) Vote(w http.ResponseWriter, r *http.Request) {
	pollID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	optionID, _ := strconv.ParseInt(r.FormValue("option_id"), 10, 64)
	user := middleware.UserFromContext(r.Context())

	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver == nil {
		http.Error(w, "No driver profile", http.StatusForbidden)
		return
	}

	poll, _ := h.polls.GetByID(r.Context(), pollID)
	if poll == nil || poll.Status != "open" {
		http.Error(w, "Poll is closed", http.StatusBadRequest)
		return
	}

	if err := h.polls.Vote(r.Context(), pollID, optionID, driver.ID, poll.AllowMultiple); err != nil {
		http.Error(w, "Failed to vote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Poll) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.tmpl.Render(w, "poll-form", map[string]any{"User": user}); err != nil {
		slog.Error("rendering poll form", "error", err)
	}
}

func (h *Poll) Create(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	closesDate := r.FormValue("closes_date")
	closesTime := r.FormValue("closes_time")
	allowMultiple := r.FormValue("allow_multiple") == "on"

	var closesAtPtr *string
	if closesDate != "" {
		closesAt := closesDate + " " + closesTime
		if closesTime == "" {
			closesAt = closesDate + " 23:59:59"
		} else {
			closesAt = closesDate + " " + closesTime + ":00"
		}
		closesAtPtr = &closesAt
	}

	pollID, err := h.polls.Create(r.Context(), title, allowMultiple, closesAtPtr)
	if err != nil {
		http.Error(w, "Failed to create poll", http.StatusInternalServerError)
		return
	}

	// Add options
	for i := 0; i < 20; i++ {
		label := r.FormValue("option_" + strconv.Itoa(i))
		if label == "" {
			continue
		}
		_ = h.polls.AddOption(r.Context(), pollID, nil, label, i)
	}

	http.Redirect(w, r, "/polls/"+strconv.FormatInt(pollID, 10), http.StatusFound)
}

func (h *Poll) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	poll, _ := h.polls.GetByID(r.Context(), id)
	if poll == nil {
		http.NotFound(w, r)
		return
	}

	newStatus := "closed"
	if poll.Status == "closed" {
		newStatus = "open"
	}
	_ = h.polls.SetStatus(r.Context(), id, newStatus)
	http.Redirect(w, r, "/polls/"+strconv.FormatInt(id, 10), http.StatusFound)
}

func (h *Poll) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())

	poll, _ := h.polls.GetByID(r.Context(), id)
	if poll == nil {
		http.NotFound(w, r)
		return
	}

	options, _ := h.polls.GetOptions(r.Context(), id)
	hasVotes, _ := h.polls.HasVotes(r.Context(), id)

	if err := h.tmpl.Render(w, "poll-edit", map[string]any{
		"User":     user,
		"Poll":     poll,
		"Options":  options,
		"HasVotes": hasVotes,
	}); err != nil {
		slog.Error("rendering poll edit", "error", err)
	}
}

func (h *Poll) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	title := r.FormValue("title")
	closesDate := r.FormValue("closes_date")
	closesTime := r.FormValue("closes_time")
	allowMultiple := r.FormValue("allow_multiple") == "on"

	var closesAtPtr *string
	if closesDate != "" {
		closesAt := closesDate + " " + closesTime
		if closesTime == "" {
			closesAt = closesDate + " 23:59:59"
		} else {
			closesAt = closesDate + " " + closesTime + ":00"
		}
		closesAtPtr = &closesAt
	}

	_ = h.polls.Update(r.Context(), id, title, allowMultiple, closesAtPtr)

	// Update options only if no votes
	hasVotes, _ := h.polls.HasVotes(r.Context(), id)
	if !hasVotes {
		type opt struct {
			TrackID *int64
			Label   string
		}
		var options []struct {
			TrackID *int64
			Label   string
		}
		for i := 0; i < 10; i++ {
			label := r.FormValue("option_" + strconv.Itoa(i))
			if label == "" {
				continue
			}
			trackIDStr := r.FormValue("track_" + strconv.Itoa(i))
			var trackID *int64
			if tid, err := strconv.ParseInt(trackIDStr, 10, 64); err == nil && tid > 0 {
				trackID = &tid
			}
			options = append(options, struct {
				TrackID *int64
				Label   string
			}{trackID, label})
		}
		if len(options) > 0 {
			_ = h.polls.ReplaceOptions(r.Context(), id, options)
		}
	}

	http.Redirect(w, r, "/polls/"+strconv.FormatInt(id, 10), http.StatusFound)
}

func (h *Poll) DeletePoll(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	_ = h.polls.Delete(r.Context(), id)
	http.Redirect(w, r, "/polls", http.StatusFound)
}
