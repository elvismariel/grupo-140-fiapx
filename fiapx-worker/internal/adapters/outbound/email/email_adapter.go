package email

import (
	"log"
	"video-processor-worker/internal/core/ports"
)

type LogEmailAdapter struct{}

func NewLogEmailAdapter() ports.EmailSender {
	return &LogEmailAdapter{}
}

func (a *LogEmailAdapter) SendEmail(to, subject, body string) error {
	log.Printf("📧 [EMAIL NOTIFICATION] To: %s | Subject: %s | Body: %s", to, subject, body)
	log.Printf("✅ Email simulation finished for %s", to)
	return nil
}
