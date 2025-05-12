package email

import (
	"context"
	"time"
)

type Email struct {
	From        string
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Data     []byte
	MIMEType string
}

type Config struct {
	DefaultSender string
	Timeout       time.Duration
}

type Provider interface {
	Send(ctx context.Context, email *Email) (messageID string, err error)
	SendTemplate(ctx context.Context, email *Email, templateID string, variables map[string]interface{}) (string, error)
}

type Service struct {
	config   Config
	provider Provider
}

func New(config Config, provider Provider) *Service {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Service{
		config:   config,
		provider: provider,
	}
}

func (s *Service) Send(ctx context.Context, email *Email) (string, error) {
	if email.From == "" {
		email.From = s.config.DefaultSender
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	return s.provider.Send(ctx, email)
}

func (s *Service) SendTemplate(ctx context.Context, email *Email, templateID string, variables map[string]interface{}) (string, error) {
	if email.From == "" {
		email.From = s.config.DefaultSender
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	return s.provider.SendTemplate(ctx, email, templateID, variables)
}
