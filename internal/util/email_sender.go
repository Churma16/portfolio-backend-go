package util

import (
	"bytes"
	_ "embed" // Penting: Jangan lupa import ini untuk fitur embed
	"fmt"
	"html/template"
	"os"

	"github.com/resend/resend-go/v2"
)

//go:embed templates/contact.html
var contactEmailTemplate string

type EmailRequest struct {
	SenderName  string
	SenderEmail string
	Content     string
}

func SendContactEmail(req EmailRequest) error {
	// 1. Parse Template HTML
	tmpl, err := template.New("contact").Parse(contactEmailTemplate)
	if err != nil {
		return fmt.Errorf("gagal parsing template email: %w", err)
	}

	// 2. Masukkan data (req) ke dalam template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, req); err != nil {
		return fmt.Errorf("gagal execute template email: %w", err)
	}

	// 3. Setup Resend
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	targetEmail := "fathanmf16@gmail.com"

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Portfolio Contact <%s>", fromEmail),
		To:      []string{targetEmail},
		Subject: fmt.Sprintf("[Pesan Baru]: %s", req.SenderName),

		// 4. Gunakan hasil render template sebagai HTML string
		Html: body.String(),

		ReplyTo: req.SenderEmail,
	}

	_, err = client.Emails.Send(params)
	if err != nil {
		fmt.Printf("Error sending email via Resend: %v\n", err)
		return err
	}

	return nil
}
