package sender

import (
	"context"
	"fmt"
	"net/smtp"

	"DelayedNotifier/internal/domain"
)

type EmailSender struct {
	smtpHost string
	smtpPort int
	from     string
	username string
	password string
}

func NewEmailSender(smtpHost string, smtpPort int, from, username, password string) *EmailSender {
	return &EmailSender{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
		username: username,
		password: password,
	}
}

func (s *EmailSender) Send(_ context.Context, recipient string, text string) error {
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)
	msg := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + recipient + "\r\n" +
			"Subject: Notification\r\n" +
			"\r\n" +
			text + "\r\n",
	)

	if err := smtp.SendMail(addr, auth, s.from, []string{recipient}, msg); err != nil {
		return fmt.Errorf("email send: %w", err)
	}

	return nil
}

func (s *EmailSender) Channel() domain.Channel {
	return domain.ChannelEmail
}
