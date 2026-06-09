package services

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"
	texttemplate "text/template"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/exnodes/hrm-api/internal/config"
)

//go:embed emailtemplates/*.html emailtemplates/*.txt
var emailTemplateFS embed.FS

// InviteEmailData is the template render context for the invite email.
type InviteEmailData struct {
	AppName   string
	FullName  string
	AcceptURL string
	ExpiresAt string
}

// EmailService renders + ships HTML/plain-text emails via SMTP. When
// SMTP_HOST is empty in the environment the service is in "disabled"
// mode: every Send returns a clear error so the caller can store it on
// the invite row's last_email_error column WITHOUT blowing up the
// outer transaction (REVISION NOTES #11).
type EmailService struct {
	cfg          *config.Config
	htmlTemplate *template.Template
	textTemplate *texttemplate.Template
}

// ErrEmailDisabled is returned when SMTP is not configured. Callers
// (InviteService.Create / Resend) catch this and persist the message
// to invites.last_email_error instead of failing the request.
var ErrEmailDisabled = errors.New("email: SMTP not configured")

// NewEmailService parses the embedded templates and constructs the
// service. Template parse errors are fatal at boot — they indicate a
// build problem, not a runtime one.
func NewEmailService(cfg *config.Config) (*EmailService, error) {
	htmlT, err := template.ParseFS(emailTemplateFS, "emailtemplates/invite.html")
	if err != nil {
		return nil, fmt.Errorf("email: parse invite.html: %w", err)
	}
	textT, err := texttemplate.ParseFS(emailTemplateFS, "emailtemplates/invite.txt")
	if err != nil {
		return nil, fmt.Errorf("email: parse invite.txt: %w", err)
	}
	return &EmailService{cfg: cfg, htmlTemplate: htmlT, textTemplate: textT}, nil
}

// IsConfigured reports whether SMTP is reachable. Callers can short-
// circuit "fire-and-forget" emails when this returns false.
func (s *EmailService) IsConfigured() bool {
	return strings.TrimSpace(s.cfg.SMTPHost) != ""
}

// SendInvite renders the invite template with data and ships it.
// Returns ErrEmailDisabled when SMTP is not configured — the caller is
// expected to record the error on the invite row.
func (s *EmailService) SendInvite(ctx context.Context, toEmail string, data InviteEmailData) error {
	if !s.IsConfigured() {
		log.Printf("email: skipped invite to %s — SMTP not configured (would-be URL: %s)", toEmail, data.AcceptURL)
		return ErrEmailDisabled
	}

	if data.AppName == "" {
		data.AppName = s.cfg.AppName
	}

	var htmlBuf bytes.Buffer
	if err := s.htmlTemplate.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("render invite.html: %w", err)
	}
	var textBuf bytes.Buffer
	if err := s.textTemplate.Execute(&textBuf, data); err != nil {
		return fmt.Errorf("render invite.txt: %w", err)
	}

	from := s.cfg.SMTPFromEmail
	if from == "" {
		from = "no-reply@" + s.cfg.AppName
	}
	fromAddr := from
	if s.cfg.SMTPFromName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, from)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromAddr)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("You're invited to %s", data.AppName))
	m.SetBody("text/plain", textBuf.String())
	m.AddAlternative("text/html", htmlBuf.String())

	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)
	d.SSL = false
	// gomail switches to STARTTLS automatically when the server advertises
	// it; SSL=true is for implicit-TLS port 465. We default to STARTTLS
	// when SMTP_USE_TLS=true; Mailpit ignores it on port 1025.
	if !s.cfg.SMTPUseTLS {
		d.TLSConfig = nil
	}

	// gomail uses net/smtp internally which doesn't honour ctx — wrap
	// with a deadline goroutine so a hung SMTP doesn't pin the request.
	done := make(chan error, 1)
	go func() {
		done <- d.DialAndSend(m)
	}()

	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
		return nil
	case <-deadline.Done():
		return fmt.Errorf("smtp send: timeout after 10s")
	}
}

// SendAnnouncementNotification sends an HTML+plaintext email notification
// for a published announcement. Uses the same gomail/SMTP pattern as
// SendInvite with a 10s timeout.
// Returns ErrEmailDisabled when SMTP is not configured.
func (s *EmailService) SendAnnouncementNotification(ctx context.Context, toEmail, title, description string) error {
	if !s.IsConfigured() {
		log.Printf("email: skipped announcement notification to %s — SMTP not configured", toEmail)
		return ErrEmailDisabled
	}

	appName := s.cfg.AppName
	if appName == "" {
		appName = "HRM"
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body style="font-family:Arial,sans-serif;line-height:1.6;color:#333;max-width:600px;margin:0 auto;padding:20px">
  <div style="background:#f8f9fa;padding:30px;border-radius:10px">
    <h2 style="color:#2563eb">%s</h2>
    <div style="line-height:1.6">%s</div>
    <hr style="border:none;border-top:1px solid #ddd;margin:20px 0">
    <p style="color:#999;font-size:12px">This is an automated message from %s.</p>
  </div>
</body></html>`, title, description, appName)

	plainText := fmt.Sprintf("%s\n\n%s\n\n-- %s", title, description, appName)

	from := s.cfg.SMTPFromEmail
	if from == "" {
		from = "no-reply@" + s.cfg.AppName
	}
	fromAddr := from
	if s.cfg.SMTPFromName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, from)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromAddr)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] New Announcement: %s", appName, title))
	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)
	d.SSL = false
	if !s.cfg.SMTPUseTLS {
		d.TLSConfig = nil
	}

	// gomail uses net/smtp internally which doesn't honour ctx — wrap
	// with a deadline goroutine so a hung SMTP doesn't pin the request.
	done := make(chan error, 1)
	go func() {
		done <- d.DialAndSend(m)
	}()

	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("smtp send announcement: %w", err)
		}
		return nil
	case <-deadline.Done():
		return fmt.Errorf("smtp send announcement: timeout after 10s")
	}
}
