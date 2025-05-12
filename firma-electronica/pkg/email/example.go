package email

import (
	"context"
	"fmt"
	"os"
	"time"
)

func Example() {
	mailgunConfig := MailgunConfig{
		Domain:        "your-domain.mailgun.org",
		APIKey:        os.Getenv("MAILGUN_API_KEY"),
		DefaultSender: "Your Name <sender@your-domain.mailgun.org>",
		Timeout:       time.Second * 30,
	}

	mailgunProvider := NewMailgunProvider(mailgunConfig)

	emailService := New(Config{
		DefaultSender: mailgunConfig.DefaultSender,
		Timeout:       time.Second * 30,
	}, mailgunProvider)

	email := &Email{
		To:      []string{"recipient@example.com"},
		Subject: "Test Email from Firma Electronica",
		Body:    "This is a test email sent from the Firma Electronica system.",
		HTMLBody: `<html>
			<body>
				<h1>Test Email</h1>
				<p>This is a <strong>test email</strong> sent from the Firma Electronica system.</p>
			</body>
		</html>`,
	}

	ctx := context.Background()
	messageID, err := emailService.Send(ctx, email)
	if err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
		return
	}

	fmt.Printf("Email sent successfully with message ID: %s\n", messageID)
}

func ExampleWithTemplate() {
	mailgunConfig := MailgunConfig{
		Domain:        "your-domain.mailgun.org",
		APIKey:        os.Getenv("MAILGUN_API_KEY"),
		DefaultSender: "Your Name <sender@your-domain.mailgun.org>",
		Timeout:       time.Second * 30,
	}

	mailgunProvider := NewMailgunProvider(mailgunConfig)

	emailService := New(Config{
		DefaultSender: mailgunConfig.DefaultSender,
		Timeout:       time.Second * 30,
	}, mailgunProvider)

	email := &Email{
		To:      []string{"recipient@example.com"},
		Subject: "Welcome to Firma Electronica",
	}

	variables := map[string]interface{}{
		"name":             "John Doe",
		"verification_url": "https://yourdomain.com/verify?token=abc123",
	}

	ctx := context.Background()
	messageID, err := emailService.SendTemplate(ctx, email, "welcome_template", variables)
	if err != nil {
		fmt.Printf("Failed to send template email: %v\n", err)
		return
	}

	fmt.Printf("Template email sent successfully with message ID: %s\n", messageID)
}

func ExampleSendDocumentToSign(recipientEmail, recipientName, documentURL, signURL string) {
	mailgunConfig := MailgunConfig{
		Domain:        "your-domain.mailgun.org",
		APIKey:        os.Getenv("MAILGUN_API_KEY"),
		DefaultSender: "Firma Electronica <no-reply@your-domain.mailgun.org>",
		Timeout:       time.Second * 30,
	}

	mailgunProvider := NewMailgunProvider(mailgunConfig)

	emailService := New(Config{
		DefaultSender: mailgunConfig.DefaultSender,
		Timeout:       time.Second * 30,
	}, mailgunProvider)

	email := &Email{
		To:      []string{recipientEmail},
		Subject: "You have a document to sign",
		Body:    fmt.Sprintf("Hello %s,\n\nYou have a document to sign. Please visit %s to sign the document.\n\nThank you,\nFirma Electronica Team", recipientName, signURL),
		HTMLBody: fmt.Sprintf(`<html>
			<body>
				<h1>Document Signature Request</h1>
				<p>Hello %s,</p>
				<p>You have a document to sign. Please click the button below to review and sign the document:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-align: center; text-decoration: none; display: inline-block; border-radius: 4px; font-weight: bold;">
						Sign Document
					</a>
				</div>
				<p>Or copy and paste this link in your browser:</p>
				<p>%s</p>
				<p>Thank you,<br>Firma Electronica Team</p>
			</body>
		</html>`, recipientName, signURL, signURL),
	}

	ctx := context.Background()
	messageID, err := emailService.Send(ctx, email)
	if err != nil {
		fmt.Printf("Failed to send document signature email: %v\n", err)
		return
	}

	fmt.Printf("Document signature email sent successfully with message ID: %s\n", messageID)
}
