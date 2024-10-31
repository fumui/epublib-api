package epublib

import (
	"context"
)

type Mail struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	From        string `json:"from"`
	To          string `json:"to"`
	Subject     string `json:"subject"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

// MailerService represents a service for managing auths.
type MailerService interface {
	// Send a mail.
	SendMail(ctx context.Context, mail Mail) error
}
