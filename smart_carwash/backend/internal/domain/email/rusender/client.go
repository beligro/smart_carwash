package rusender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const apiURL = "https://api.rusender.ru/api/v1/external-mails/send"

// Client клиент RuSender для отправки писем
type Client struct {
	apiKey     string
	fromEmail  string
	httpClient *http.Client
}

// NewClient создаёт клиент RuSender
func NewClient(apiKey, fromEmail string) *Client {
	return &Client{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendEmailRequest тело запроса по документации RuSender
type SendEmailRequest struct {
	IdempotencyKey string   `json:"idempotencyKey,omitempty"`
	Mail           MailBody `json:"mail"`
}

// MailBody структура письма
type MailBody struct {
	To      MailAddress `json:"to"`
	From    MailAddress `json:"from"`
	Subject string      `json:"subject"`
	HTML    string      `json:"html,omitempty"`
	Text    string      `json:"text,omitempty"`
}

// MailAddress email и имя
type MailAddress struct {
	Email string `json:"email"`
}

// Send отправляет письмо. subject и body (html или text) обязательны.
func (c *Client) Send(ctx context.Context, toEmail, subject, htmlBody, textBody string) error {
	if c.apiKey == "" {
		return fmt.Errorf("rusender: API key not configured")
	}
	if toEmail == "" {
		return fmt.Errorf("rusender: recipient email required")
	}
	if subject == "" {
		return fmt.Errorf("rusender: subject required")
	}
	if htmlBody == "" && textBody == "" {
		return fmt.Errorf("rusender: html or text body required")
	}
	body := MailBody{
		To:      MailAddress{Email: toEmail},
		From:    MailAddress{Email: c.fromEmail},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}
	if body.HTML == "" {
		body.HTML = textBody
	}
	if body.Text == "" {
		body.Text = htmlBody
	}
	payload := SendEmailRequest{
		IdempotencyKey: uuid.New().String(),
		Mail:           body,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rusender marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("rusender request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errBody struct {
			Message    string `json:"message"`
			StatusCode int    `json:"statusCode"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Message != "" {
			return fmt.Errorf("rusender: %s (status %d)", errBody.Message, resp.StatusCode)
		}
		return fmt.Errorf("rusender: status %d", resp.StatusCode)
	}
	return nil
}
