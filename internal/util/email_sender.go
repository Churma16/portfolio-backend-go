package util

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailRequest struct {
	SenderName  string
	SenderEmail string
	Content     string
}

func SendContactEmail(req EmailRequest) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL") // onboarding@resend.dev atau verified domain kamu

	// Email tujuan (Email kamu sendiri)
	targetEmail := "fathanmf16@gmail.com"

	client := resend.NewClient(apiKey)

	// Format HTML Body
	htmlContent := fmt.Sprintf(`
		<h3>Pesan Baru dari Portfolio</h3>
		<p><strong>Nama:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<br>
		<p><strong>Pesan:</strong></p>
		<p>%s</p>
	`, req.SenderName, req.SenderEmail, req.Content)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("PORTO CONTACT <%s>", fromEmail),
		To:      []string{targetEmail},
		Subject: fmt.Sprintf("Pesan Baru dari: %s", req.SenderName),
		Html:    htmlContent,

		// FITUR PENTING: Reply-To
		// Jadi pas kamu klik "Reply" di Gmail, dia lgsg balas ke email pengirim (bukan ke onboarding@resend.dev)
		ReplyTo: req.SenderEmail,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		fmt.Printf("Error sending email via Resend: %v\n", err)
		return err
	}

	return nil
}
