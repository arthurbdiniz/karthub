package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/karthub/karthub/internal/config"
	"github.com/karthub/karthub/internal/handler"
	"github.com/karthub/karthub/internal/mail"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/session"
	"github.com/karthub/karthub/internal/templates"
)

type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
	router *chi.Mux
}

func New(cfg *config.Config, logger *slog.Logger, db *sql.DB) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
	s.setupRouter()
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Logger(s.logger))
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Compress(5))

	// Templates
	tmpl := templates.New()

	// Repositories
	userRepo := repository.NewUserRepository(s.db)
	tokenRepo := repository.NewMagicTokenRepository(s.db)
	driverRepo := repository.NewDriverRepository(s.db)
	trackRepo := repository.NewTrackRepository(s.db)
	eventRepo := repository.NewEventRepository(s.db)
	bookingRepo := repository.NewBookingRepository(s.db)
	photoRepo := repository.NewPhotoRepository(s.db)
	resultRepo := repository.NewResultRepository(s.db)
	feedbackRepo := repository.NewFeedbackRepository(s.db)

	// Session manager
	sessionMgr := session.NewManager(s.db, s.cfg.Session)

	// Mail sender
	mailer := mail.NewSender(s.cfg.Mail, s.logger)

	// Auth middleware
	authMw := middleware.NewAuth(sessionMgr, s.db)

	// Handlers
	authHandler := handler.NewAuth(userRepo, tokenRepo, sessionMgr, mailer, tmpl, s.cfg.BaseURL)
	dashHandler := handler.NewDashboard(eventRepo, bookingRepo, driverRepo, resultRepo, userRepo, tmpl)
	driverHandler := handler.NewDriver(driverRepo, resultRepo, tmpl)
	trackHandler := handler.NewTrack(trackRepo, tmpl)
	eventHandler := handler.NewEvent(eventRepo, bookingRepo, trackRepo, driverRepo, photoRepo, resultRepo, tmpl)
	bookingHandler := handler.NewBooking(bookingRepo, eventRepo, tmpl)
	mediaHandler := handler.NewMedia(photoRepo, driverRepo, "data/uploads")
	resultHandler := handler.NewResult(resultRepo)
	feedbackHandler := handler.NewFeedback(feedbackRepo, driverRepo, tmpl)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(templates.StaticFS()))))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/login", authHandler.LoginPage)
		r.Post("/login", authHandler.SendMagicLink)
		r.Get("/auth/verify", authHandler.Verify)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(authMw.Required)
		r.Use(authMw.ProfileRequired)

		r.Get("/", dashHandler.Index)
		r.Post("/logout", authHandler.Logout)

		// Uploaded media (access controlled)
		r.Get("/uploads/{filename}", mediaHandler.ServeFile)

		// First-time profile setup
		r.Get("/setup", driverHandler.SetupForm)
		r.Post("/setup", driverHandler.Setup)

		// Drivers
		r.Route("/drivers", func(r chi.Router) {
			r.Get("/", driverHandler.List)
			r.Get("/new", driverHandler.NewForm)
			r.Post("/new", driverHandler.CreateProfile)
			r.Get("/me/edit", driverHandler.EditMyProfile)
			r.Post("/me", driverHandler.UpdateMyProfile)
			r.Get("/{id}", driverHandler.Show)
		})

		// Tracks
		r.Route("/tracks", func(r chi.Router) {
			r.Get("/", trackHandler.List)
		})

		// Events
		r.Route("/events", func(r chi.Router) {
			r.Get("/", eventHandler.List)
			r.Get("/{id}", eventHandler.Show)
			r.Post("/{eventID}/photos", mediaHandler.Upload)
			r.Post("/{eventID}/photos/{id}/up", mediaHandler.MoveUp)
			r.Post("/{eventID}/photos/{id}/down", mediaHandler.MoveDown)
			r.Post("/{eventID}/photos/reorder", mediaHandler.Reorder)
			r.Post("/{eventID}/results", resultHandler.Save)
			r.Post("/{eventID}/results/reset", resultHandler.Delete)
			r.Get("/{eventID}/results", resultHandler.Table)
			r.Delete("/photos/{id}", mediaHandler.Delete)
		})

		// Bookings
		r.Route("/bookings", func(r chi.Router) {
			r.Post("/", bookingHandler.Create)
			r.Delete("/{id}", bookingHandler.Cancel)
			r.Post("/{id}/confirm", bookingHandler.Confirm)
			r.Post("/{id}/unconfirm", bookingHandler.Unconfirm)
			r.Post("/{id}/waitlist", bookingHandler.MoveToWaitlist)
			r.Post("/{id}/promote", bookingHandler.Promote)
			r.Delete("/{id}/admin", bookingHandler.Remove)
		})

		// Feedback
		r.Get("/feedback", feedbackHandler.Form)
		r.Post("/feedback", feedbackHandler.Submit)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(authMw.AdminOrOrganizer)
			r.Get("/", dashHandler.Admin)
			r.Route("/drivers", func(r chi.Router) {
				r.Get("/new", driverHandler.NewForm)
				r.Post("/", driverHandler.CreateDriver)
				r.Get("/{id}/edit", driverHandler.EditForm)
				r.Post("/{id}", driverHandler.Update)
				r.Delete("/{id}", driverHandler.Delete)
			})
			r.Route("/tracks", func(r chi.Router) {
				r.Get("/new", trackHandler.NewForm)
				r.Post("/", trackHandler.CreateTrack)
				r.Get("/{id}/edit", trackHandler.EditForm)
				r.Post("/{id}", trackHandler.Update)
				r.Delete("/{id}", trackHandler.Delete)
			})
			r.Route("/events", func(r chi.Router) {
				r.Get("/new", eventHandler.NewForm)
				r.Post("/", eventHandler.CreateEvent)
				r.Get("/{id}/edit", eventHandler.EditForm)
				r.Post("/{id}", eventHandler.Update)
				r.Post("/{id}/status", eventHandler.ToggleStatus)
				r.Delete("/{id}", eventHandler.Delete)
			})

			// Admin-only: user role management
			r.Route("/users", func(r chi.Router) {
				r.Use(authMw.AdminOnly)
				r.Get("/", dashHandler.Users)
				r.Post("/{id}/role", dashHandler.SetRole)
			})

			// Feedback (visible to admin + organizer)
			r.Get("/feedback", feedbackHandler.List)
		})
	})

	s.router = r
}
