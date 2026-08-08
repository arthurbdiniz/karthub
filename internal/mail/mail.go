package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/karthub/karthub/internal/config"
)

type Sender struct {
	cfg    config.MailConfig
	client *http.Client
	logger *slog.Logger
}

func NewSender(cfg config.MailConfig, logger *slog.Logger) *Sender {
	return &Sender{
		cfg:    cfg,
		client: &http.Client{},
		logger: logger,
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *Sender) SendMagicLink(ctx context.Context, to, link string) error {
	if s.cfg.DevMode {
		s.logger.Info("magic link (dev mode)", "to", to, "link", link)
		fmt.Fprintf(os.Stderr, "\n\033[1;36m✉ Magic link for %s:\033[0m\n  %s\n\n", to, link)
		return nil
	}

	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="color: #1a1a1a;">🏎️ KartHub Login</h2>
			<p style="color: #4a4a4a; font-size: 16px;">Click the button below to sign in. This link expires in 15 minutes.</p>
			<a href="%s" style="display: inline-block; background: #2563eb; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none; font-weight: 600; margin: 16px 0;">Sign In</a>
			<p style="color: #9a9a9a; font-size: 13px;">If you didn't request this, you can safely ignore this email.</p>
		</div>
	`, link)

	return s.send(ctx, to, "Your KartHub login link", html)
}

func (s *Sender) send(ctx context.Context, to, subject, html string) error {
	from := fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromAddress)

	payload := sendRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		var errResp map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			s.logger.Error("decoding resend error response", "error", err)
		}
		s.logger.Error("resend API error", "status", resp.StatusCode, "response", errResp)
		return fmt.Errorf("resend API returned %d", resp.StatusCode)
	}

	s.logger.Info("email sent", "to", to, "subject", subject)
	return nil
}
