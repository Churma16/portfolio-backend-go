package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type NotificationPayload struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

var notifierHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SendNotification dispatches contact form notifications to portfolio-notifier microservice
func SendNotification(name, email, message, source string) error {
	notifierURL := os.Getenv("NOTIFIER_URL")
	if notifierURL == "" {
		notifierURL = "http://portfolio-notifier:8080/api/notify"
	}

	payload := NotificationPayload{
		Name:    name,
		Email:   email,
		Message: message,
		Source:  source,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, notifierURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := notifierHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier returned status code %d", resp.StatusCode)
	}

	return nil
}

// EmailRequest preserved for backwards compatibility
type EmailRequest struct {
	SenderName  string
	SenderEmail string
	Content     string
}

// SendContactEmail delegates to SendNotification
func SendContactEmail(req EmailRequest) error {
	return SendNotification(req.SenderName, req.SenderEmail, req.Content, "go")
}
