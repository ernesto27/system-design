package email

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// readCloserWrapper wraps an io.Reader to provide io.ReadCloser interface
type readCloserWrapper struct {
	reader io.Reader
}

func (r *readCloserWrapper) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *readCloserWrapper) Close() error {
	// Check if the underlying reader is also a closer
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// newReadCloserWrapper converts an io.Reader to an io.ReadCloser
func newReadCloserWrapper(reader io.Reader) io.ReadCloser {
	return &readCloserWrapper{reader: reader}
}

type MailgunConfig struct {
	Domain        string
	APIKey        string
	DefaultSender string
	Timeout       time.Duration
}

type MailgunProvider struct {
	client *mailgun.MailgunImpl
	config MailgunConfig
}

func NewMailgunProvider(config MailgunConfig) *MailgunProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	mg := mailgun.NewMailgun(config.Domain, config.APIKey)
	return &MailgunProvider{
		client: mg,
		config: config,
	}
}

func (p *MailgunProvider) Send(ctx context.Context, email *Email) (string, error) {
	emails := make([]string, len(email.To))
	for _, to := range email.To {
		emails = append(emails, "name <"+to+">")
	}

	message := p.client.NewMessage(
		email.From,
		email.Subject,
		email.Body,
		emails...,
	)

	if email.HTMLBody != "" {
		message.SetHtml(email.HTMLBody)
	}

	for _, attachment := range email.Attachments {
		message.AddBufferAttachment(attachment.Filename, attachment.Data)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	resp, id, err := p.client.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send email via Mailgun: %w", err)
	}

	fmt.Printf("Mailgun response: %s, ID: %s\n", resp, id)

	return id, nil
}

func (p *MailgunProvider) SendTemplate(ctx context.Context, email *Email, templateID string, variables map[string]interface{}) (string, error) {
	message := p.client.NewMessage(
		email.From,
		email.Subject,
		email.Body,
		email.To...,
	)

	message.SetTemplate(templateID)

	for k, v := range variables {
		if err := message.AddTemplateVariable(k, v); err != nil {
			return "", fmt.Errorf("failed to add template variable %s: %w", k, err)
		}
	}

	for _, attachment := range email.Attachments {
		message.AddBufferAttachment(attachment.Filename, attachment.Data)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	resp, id, err := p.client.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send template email via Mailgun: %w", err)
	}

	fmt.Printf("Mailgun template response: %s, ID: %s\n", resp, id)

	return id, nil
}

func (p *MailgunProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	mv := mailgun.NewEmailValidator(p.config.APIKey)

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	resp, err := mv.ValidateEmail(ctx, email, false)
	if err != nil {
		return false, fmt.Errorf("failed to validate email: %w", err)
	}

	return resp.IsValid, nil
}

func (p *MailgunProvider) SendWithAttachmentReader(ctx context.Context, email *Email, attachments map[string]io.Reader) (string, error) {
	message := p.client.NewMessage(
		email.From,
		email.Subject,
		email.Body,
		email.To...,
	)

	if email.HTMLBody != "" {
		message.SetHtml(email.HTMLBody)
	}

	for name, reader := range attachments {
		// Wrap the io.Reader with our ReadCloser wrapper
		readCloser := newReadCloserWrapper(reader)
		message.AddReaderAttachment(name, readCloser)
	}

	for _, attachment := range email.Attachments {
		message.AddBufferAttachment(attachment.Filename, attachment.Data)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	resp, id, err := p.client.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send email with reader attachments via Mailgun: %w", err)
	}

	fmt.Printf("Mailgun response: %s, ID: %s\n", resp, id)

	return id, nil
}
