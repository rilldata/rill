package email

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"

	"go.uber.org/zap"
)

type Sender interface {
	Send(toEmail, toName, subject, body string) error
}

type SMTPOptions struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BCC          string
}

type smtpSender struct {
	opts *SMTPOptions
}

func NewSMTPSender(opts *SMTPOptions) (Sender, error) {
	if opts.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP server password is required")
	}

	_, err := mail.ParseAddress(opts.FromEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid sender email address %q", opts.FromEmail)
	}

	if opts.BCC != "" {
		_, err := mail.ParseAddress(opts.BCC)
		if err != nil {
			return nil, fmt.Errorf("invalid bcc email address %q", opts.BCC)
		}
	}

	return &smtpSender{opts: opts}, nil
}

func (s *smtpSender) Send(toEmail, toName, subject, body string) error {
	// Compose the email message
	from := mail.Address{Name: s.opts.FromName, Address: s.opts.FromEmail}
	to := mail.Address{Name: toName, Address: toEmail}
	message := []byte("From: " + from.String() + "\r\n" +
		"To: " + to.String() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		body + "\r\n",
	)

	// Build recipients list
	recipients := []string{toEmail}
	if s.opts.BCC != "" {
		recipients = append(recipients, s.opts.BCC)
	}

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", s.opts.SMTPUsername, s.opts.SMTPPassword, s.opts.SMTPHost)
	err := smtp.SendMail(s.opts.SMTPHost+":"+strconv.Itoa(s.opts.SMTPPort), auth, from.Address, recipients, message)
	if err != nil {
		return err
	}

	return nil
}

type consoleSender struct {
	logger    *zap.Logger
	fromEmail string
	fromName  string
}

func NewConsoleSender(logger *zap.Logger, fromEmail, fromName string) (Sender, error) {
	_, err := mail.ParseAddress(fromEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid sender email address %q", fromEmail)
	}

	return &consoleSender{fromEmail: fromEmail, fromName: fromName}, nil
}

func (s *consoleSender) Send(toEmail, toName, subject, body string) error {
	s.logger.Info("email sent",
		zap.String("from_email", s.fromEmail),
		zap.String("from_name", s.fromName),
		zap.String("to_email", toEmail),
		zap.String("to_name", toName),
		zap.String("subject", subject),
		zap.String("body", body),
	)
	return nil
}
