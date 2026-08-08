package handler

import (
	"net/http"

	"github.com/karthub/karthub/internal/mail"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/session"
	"github.com/karthub/karthub/internal/templates"
)

type Auth struct {
	users    *repository.UserRepository
	tokens   *repository.MagicTokenRepository
	sessions *session.Manager
	mail     *mail.Sender
	tmpl     *templates.Templates
	baseURL  string
}

func NewAuth(
	users *repository.UserRepository,
	tokens *repository.MagicTokenRepository,
	sessions *session.Manager,
	mail *mail.Sender,
	tmpl *templates.Templates,
	baseURL string,
) *Auth {
	return &Auth{
		users:    users,
		tokens:   tokens,
		sessions: sessions,
		mail:     mail,
		tmpl:     tmpl,
		baseURL:  baseURL,
	}
}

func (h *Auth) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Render(w, "login", nil)
}

func (h *Auth) SendMagicLink(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		h.tmpl.Render(w, "login", map[string]any{"Error": "Email is required"})
		return
	}

	token, err := h.tokens.Create(r.Context(), email)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	link := h.baseURL + "/auth/verify?token=" + token

	if err := h.mail.SendMagicLink(r.Context(), email, link); err != nil {
		h.tmpl.Render(w, "login", map[string]any{"Error": "Failed to send email. Please try again."})
		return
	}

	h.tmpl.RenderPartial(w, "check-email", map[string]any{"Email": email})
}

func (h *Auth) Verify(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		h.tmpl.Render(w, "login", map[string]any{"Error": "Invalid or expired link"})
		return
	}

	email, err := h.tokens.Consume(r.Context(), tokenStr)
	if err != nil {
		h.tmpl.Render(w, "login", map[string]any{"Error": "Invalid or expired link"})
		return
	}

	// Find or create user
	user, err := h.users.GetByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		user, err = h.users.CreateFromEmail(r.Context(), email)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}

	if err := h.sessions.Create(w, user.ID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Destroy(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}
