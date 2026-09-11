package email

// import (
// 	"bytes"
// 	"fmt"
// 	"html/template"

// 	"github.com/resend/resend-go/v3"
// 	"github.com/riazahmedshah/stayz/internal/config"
// )

// type Client struct {
// 	client *resend.Client
// }

// func NewClient(cfg config.IntegrationConfig) *Client {
// 	return &Client{
// 		client: resend.NewClient(cfg.ResendAPIKey),
// 	}
// }

// func (c *Client) SendEmail(to, subject string, templateName string, data map[string]any) error {
// 	templatePath := fmt.Sprintf("%s/%s.html", "templates/email", templateName)

// 	tmp, err := template.ParseFiles(templatePath)

// 	if err != nil {
// 		return fmt.Errorf("failed to parse email template %s", templateName)
// 	}

// 	var body bytes.Buffer
// 	if err := tmp.Execute(&body, data); err != nil {
// 		return fmt.Errorf("failed to execute email template %s", templateName)
// 	}

// 	params := &resend.SendEmailRequest{
// 		From:    fmt.Sprintf("%s <%s>", "stayz", "confirmation@resend.dev"),
// 		To:      []string{to},
// 		Subject: subject,
// 		Html:    body.String(),
// 	}

// 	_, err = c.client.Emails.Send(params)
// 	if err != nil {
// 		return fmt.Errorf("failed to send email: %w", err)
// 	}

// 	return nil
// }
