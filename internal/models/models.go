package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MagicToken struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type Driver struct {
	ID          int64     `json:"id"`
	UserID      *int64    `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	Nickname    *string   `json:"nickname,omitempty"`
	Avatar      *string   `json:"avatar,omitempty"`
	CountryCode *string   `json:"country_code,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Track struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	City         string    `json:"city"`
	LengthMeters *int      `json:"length_meters,omitempty"`
	Indoor       bool      `json:"indoor"`
	LocationURL  *string   `json:"location_url,omitempty"`
	MapEmbed     *string   `json:"map_embed,omitempty"`
	Website      *string   `json:"website,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Championship struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Season       string    `json:"season"`
	PointsSystem string    `json:"points_system"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Event struct {
	ID               int64     `json:"id"`
	ChampionshipID   *int64    `json:"championship_id,omitempty"`
	TrackID          int64     `json:"track_id"`
	Date             string    `json:"date"`
	Time             *string   `json:"time,omitempty"`
	DurationMinutes  int       `json:"duration_minutes"`
	MaxDrivers       int       `json:"max_drivers"`
	EntryFeeCents    int       `json:"entry_fee_cents"`
	PaymentInfo      *string   `json:"payment_info,omitempty"`
	OrganizerContact *string   `json:"organizer_contact,omitempty"`
	Status           string    `json:"status"`
	Notes            *string   `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Booking struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"event_id"`
	DriverID  int64     `json:"driver_id"`
	Status    string    `json:"status"`
	Position  *int      `json:"position,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Result struct {
	ID             int64     `json:"id"`
	EventID        int64     `json:"event_id"`
	DriverID       int64     `json:"driver_id"`
	Position       int       `json:"position"`
	BestLapTime    *string   `json:"best_lap_time,omitempty"`
	FastestLap     bool      `json:"fastest_lap"`
	DNF            bool      `json:"dnf"`
	PenaltySeconds int       `json:"penalty_seconds"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type ChampionshipPoints struct {
	ID             int64     `json:"id"`
	ChampionshipID int64     `json:"championship_id"`
	EventID        int64     `json:"event_id"`
	DriverID       int64     `json:"driver_id"`
	Points         int       `json:"points"`
	CreatedAt      time.Time `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Data      *string   `json:"data,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EventPhoto struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	DriverID     int64     `json:"driver_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type Feedback struct {
	ID        int64     `json:"id"`
	DriverID  int64     `json:"driver_id"`
	EventID   *int64    `json:"event_id,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
