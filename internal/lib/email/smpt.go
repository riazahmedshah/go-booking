package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/riazahmedshah/stayz/internal/config"
)

type SMTPClient struct {
	smtpHost string
	smtpPort string
}

func NewSMTPClient(cfg *config.Config) *SMTPClient {
	return &SMTPClient{
		smtpHost: cfg.Integration.SMTPHost,
		smtpPort: cfg.Integration.SMTPPort,
	}
}

func (c *SMTPClient) SendEmail(to, subject, templateName string, data map[string]any) error {
	templatePath := fmt.Sprintf("%s/%s.html", "templates/emails", templateName)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template %s", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template %s", templateName)
	}

	from := "confirmation@stayz.co.in"
	msg := fmt.Appendf(nil, "From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body.String())

	// Mailpit doesn't require authentication (nil auth)
	addr := fmt.Sprintf("%s:%s", c.smtpHost, c.smtpPort)
	if err := smtp.SendMail(addr, nil, from, []string{to}, msg); err != nil {
		return fmt.Errorf("failed to send email via Mailpit: %w", err)
	}

	return nil
}
