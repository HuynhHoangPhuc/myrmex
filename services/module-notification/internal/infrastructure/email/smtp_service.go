package email

import (
	"fmt"
	"time"

	gomail "github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

// SMTPConfig holds SMTP connection parameters.
type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// SMTPService wraps go-mail for HTML email delivery.
// A nil SMTPService is a valid no-op (used when SMTP not configured).
type SMTPService struct {
	cfg SMTPConfig
	log *zap.Logger
}

// NewSMTPService returns an SMTPService if cfg.Host is set, otherwise nil (no-op mode).
func NewSMTPService(cfg SMTPConfig, log *zap.Logger) *SMTPService {
	if cfg.Host == "" {
		log.Warn("SMTP host not configured, email delivery disabled")
		return nil
	}
	return &SMTPService{cfg: cfg, log: log}
}

// Send delivers an HTML email to a single recipient.
func (s *SMTPService) Send(to, subject, htmlBody string) error {
	if s == nil {
		return nil // no-op
	}

	msg := gomail.NewMsg()
	if err := msg.FromFormat(s.cfg.FromName, s.cfg.FromEmail); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, htmlBody)

	opts := []gomail.Option{
		gomail.WithPort(s.cfg.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(s.cfg.Username),
		gomail.WithPassword(s.cfg.Password),
		gomail.WithTimeout(15 * time.Second),
	}
	// TLS negotiation: use STARTTLS for port 587, implicit TLS for 465
	if s.cfg.Port == 465 {
		opts = append(opts, gomail.WithSSLPort(false))
	} else {
		opts = append(opts, gomail.WithTLSPortPolicy(gomail.TLSMandatory))
	}

	client, err := gomail.NewClient(s.cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("send email to %s: %w", to, err)
	}

	s.log.Debug("email sent", zap.String("to", to), zap.String("subject", subject))
	return nil
}
