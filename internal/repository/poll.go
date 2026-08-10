package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PollRepository struct {
	db *sql.DB
}

func NewPollRepository(db *sql.DB) *PollRepository {
	return &PollRepository{db: db}
}

type Poll struct {
	ID            int64
	Title         string
	Status        string
	AllowMultiple bool
	ClosesAt      *string
	CreatedAt     string
}

type PollOption struct {
	ID        int64
	PollID    int64
	TrackID   *int64
	Label     string
	SortOrder int
}

type PollVoter struct {
	DriverID       int64
	DriverNickname *string
	DriverAvatar   *string
}

type PollOptionWithVotes struct {
	PollOption
	VoteCount int
	Voters    []PollVoter
}

func (r *PollRepository) Create(ctx context.Context, title string, allowMultiple bool, closesAt *string) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO polls (title, allow_multiple, closes_at) VALUES (?, ?, ?)", title, allowMultiple, closesAt)
	if err != nil {
		return 0, fmt.Errorf("creating poll: %w", err)
	}
	return result.LastInsertId()
}

func (r *PollRepository) AddOption(ctx context.Context, pollID int64, trackID *int64, label string, order int) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO poll_options (poll_id, track_id, label, sort_order) VALUES (?, ?, ?, ?)",
		pollID, trackID, label, order)
	return err
}

func (r *PollRepository) GetByID(ctx context.Context, id int64) (*Poll, error) {
	var p Poll
	err := r.db.QueryRowContext(ctx,
		"SELECT id, title, status, allow_multiple, closes_at, created_at FROM polls WHERE id = ?", id).Scan(
		&p.ID, &p.Title, &p.Status, &p.AllowMultiple, &p.ClosesAt, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying poll: %w", err)
	}
	// Auto-close if past deadline
	if p.Status == "open" && p.ClosesAt != nil {
		now := time.Now().UTC().Format("2006-01-02 15:04:05")
		if now > *p.ClosesAt {
			_, _ = r.db.ExecContext(ctx, "UPDATE polls SET status = 'closed' WHERE id = ?", id)
			p.Status = "closed"
		}
	}
	return &p, nil
}

func (r *PollRepository) ListActive(ctx context.Context) ([]Poll, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, status, allow_multiple, closes_at, created_at FROM polls ORDER BY CASE WHEN status = 'open' THEN 0 ELSE 1 END, created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("querying polls: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var polls []Poll
	for rows.Next() {
		var p Poll
		if err := rows.Scan(&p.ID, &p.Title, &p.Status, &p.AllowMultiple, &p.ClosesAt, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning poll: %w", err)
		}
		polls = append(polls, p)
	}
	return polls, rows.Err()
}

func (r *PollRepository) GetOptions(ctx context.Context, pollID int64) ([]PollOptionWithVotes, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT po.id, po.poll_id, po.track_id, po.label, po.sort_order,
			(SELECT COUNT(*) FROM poll_votes pv WHERE pv.option_id = po.id) as vote_count
		FROM poll_options po
		WHERE po.poll_id = ?
		ORDER BY po.sort_order
	`, pollID)
	if err != nil {
		return nil, fmt.Errorf("querying poll options: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var options []PollOptionWithVotes
	for rows.Next() {
		var o PollOptionWithVotes
		if err := rows.Scan(&o.ID, &o.PollID, &o.TrackID, &o.Label, &o.SortOrder, &o.VoteCount); err != nil {
			return nil, fmt.Errorf("scanning option: %w", err)
		}
		options = append(options, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load voters for each option
	for i := range options {
		voters, _ := r.getVoters(ctx, options[i].ID)
		options[i].Voters = voters
	}

	return options, nil
}

func (r *PollRepository) getVoters(ctx context.Context, optionID int64) ([]PollVoter, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.nickname, d.avatar
		FROM poll_votes pv
		JOIN drivers d ON d.id = pv.driver_id
		WHERE pv.option_id = ?
	`, optionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var voters []PollVoter
	for rows.Next() {
		var v PollVoter
		if err := rows.Scan(&v.DriverID, &v.DriverNickname, &v.DriverAvatar); err != nil {
			return nil, err
		}
		voters = append(voters, v)
	}
	return voters, rows.Err()
}

func (r *PollRepository) Vote(ctx context.Context, pollID, optionID, driverID int64, allowMultiple bool) error {
	if !allowMultiple {
		// Single vote: remove any existing vote then insert
		_, _ = r.db.ExecContext(ctx, "DELETE FROM poll_votes WHERE poll_id = ? AND driver_id = ?", pollID, driverID)
		_, err := r.db.ExecContext(ctx,
			"INSERT INTO poll_votes (poll_id, option_id, driver_id) VALUES (?, ?, ?)",
			pollID, optionID, driverID)
		return err
	}
	// Multiple: toggle vote on this option
	var exists int
	_ = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM poll_votes WHERE poll_id = ? AND option_id = ? AND driver_id = ?",
		pollID, optionID, driverID).Scan(&exists)
	if exists > 0 {
		_, err := r.db.ExecContext(ctx,
			"DELETE FROM poll_votes WHERE poll_id = ? AND option_id = ? AND driver_id = ?",
			pollID, optionID, driverID)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO poll_votes (poll_id, option_id, driver_id) VALUES (?, ?, ?)",
		pollID, optionID, driverID)
	return err
}

func (r *PollRepository) GetUserVotes(ctx context.Context, pollID, driverID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT option_id FROM poll_votes WHERE poll_id = ? AND driver_id = ?", pollID, driverID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PollRepository) SetStatus(ctx context.Context, pollID int64, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE polls SET status = ? WHERE id = ?", status, pollID)
	return err
}

func (r *PollRepository) Update(ctx context.Context, id int64, title string, allowMultiple bool, closesAt *string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE polls SET title = ?, allow_multiple = ?, closes_at = ? WHERE id = ?", title, allowMultiple, closesAt, id)
	return err
}

func (r *PollRepository) HasVotes(ctx context.Context, pollID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM poll_votes WHERE poll_id = ?", pollID).Scan(&count)
	return count > 0, err
}

func (r *PollRepository) ReplaceOptions(ctx context.Context, pollID int64, options []struct {
	TrackID *int64
	Label   string
}) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM poll_options WHERE poll_id = ?", pollID)
	if err != nil {
		return err
	}
	for i, o := range options {
		_, err := r.db.ExecContext(ctx,
			"INSERT INTO poll_options (poll_id, track_id, label, sort_order) VALUES (?, ?, ?, ?)",
			pollID, o.TrackID, o.Label, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PollRepository) Delete(ctx context.Context, id int64) error {
	_, _ = r.db.ExecContext(ctx, "DELETE FROM poll_votes WHERE poll_id = ?", id)
	_, _ = r.db.ExecContext(ctx, "DELETE FROM poll_options WHERE poll_id = ?", id)
	_, err := r.db.ExecContext(ctx, "DELETE FROM polls WHERE id = ?", id)
	return err
}

type PollListItem struct {
	Poll
	VoteCount int
}

func (r *PollRepository) ListWithCounts(ctx context.Context) ([]PollListItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.status, p.allow_multiple, p.closes_at, p.created_at,
			(SELECT COUNT(*) FROM poll_votes pv WHERE pv.poll_id = p.id) as vote_count
		FROM polls p
		ORDER BY CASE WHEN p.status = 'open' THEN 0 ELSE 1 END, p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying polls with counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var polls []PollListItem
	for rows.Next() {
		var p PollListItem
		if err := rows.Scan(&p.ID, &p.Title, &p.Status, &p.AllowMultiple, &p.ClosesAt, &p.CreatedAt, &p.VoteCount); err != nil {
			return nil, fmt.Errorf("scanning poll: %w", err)
		}
		polls = append(polls, p)
	}
	return polls, rows.Err()
}
