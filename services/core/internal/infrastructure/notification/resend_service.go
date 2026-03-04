package notification

import (
	"fmt"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

// EmailService sends transactional emails via Resend.
type EmailService struct {
	client    *resend.Client
	fromAddr  string
	log       *zap.Logger
}

// NewEmailService constructs an EmailService. Returns nil when apiKey is empty (email disabled).
func NewEmailService(apiKey, fromAddr string, log *zap.Logger) *EmailService {
	if apiKey == "" {
		log.Info("resend api key not set — email notifications disabled")
		return nil
	}
	return &EmailService{
		client:   resend.NewClient(apiKey),
		fromAddr: fromAddr,
		log:      log,
	}
}

// Send delivers an email. Errors are logged but not propagated — graceful degradation.
func (s *EmailService) Send(to, notifType string, templateData any) {
	subject, htmlBody, err := RenderEmail(notifType, templateData)
	if err != nil {
		s.log.Warn("email template render failed",
			zap.String("type", notifType), zap.Error(err))
		return
	}

	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}
	if _, err := s.client.Emails.Send(params); err != nil {
		s.log.Warn("resend send failed",
			zap.String("to", to),
			zap.String("type", notifType),
			zap.Error(err))
	}
}

// GetUserEmail is a helper to retrieve an email address from core.users by ID.
// Implemented as a simple closure set at wire-up to avoid circular imports.
var GetUserEmail func(userID string) (string, error)

// userEmailFor wraps GetUserEmail with a friendly error message.
func userEmailFor(userID string) (string, error) {
	if GetUserEmail == nil {
		return "", fmt.Errorf("GetUserEmail not wired")
	}
	return GetUserEmail(userID)
}
