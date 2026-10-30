package mailer

import (
	"context"
	"fmt"
	"os"

	"github.com/wneessen/go-mail"

	"github.com/xoxyn/semul/internal/domain"
)

// SMTP sends emails through an SMTP server (e.g. Gmail with an app password).
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// FromEnv builds a Gmail SMTP config from SMTP_USERNAME / SMTP_PASSWORD.
// SMTP_FROM is optional and defaults to the username.
func FromEnv() (SMTP, error) {
	s := SMTP{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
	if s.Username == "" || s.Password == "" {
		return SMTP{}, fmt.Errorf("SMTP_USERNAME and SMTP_PASSWORD must be set")
	}
	if s.From == "" {
		s.From = s.Username
	}
	return s, nil
}

// Send delivers one job's draft to the prospect. Its signature matches
// tui.SendFunc, so you can pass the method value directly: tui.Run(jobs, smtp.Send).
func (s SMTP) Send(ctx context.Context, job domain.EmailJob) error {
	msg := mail.NewMsg()

	if err := msg.From(s.From); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(job.Email); err != nil {
		return fmt.Errorf("set to %q: %w", job.Email, err)
	}
	msg.Subject(job.Draft.Subject)
	msg.SetBodyString(mail.TypeTextPlain, job.Draft.Body)

	client, err := mail.NewClient(
		s.Host,
		mail.WithPort(s.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.Username),
		mail.WithPassword(s.Password),
	)
	if err != nil {
		return fmt.Errorf("create mail client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send to %s: %w", job.Email, err)
	}
	return nil
}
