package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"

	"go.uber.org/zap"
)

type Sender interface {
	Send(toEmail, toName, subject, body string, attachments []Attachment) error
}

// Attachment is a file attached to an email.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
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

func (s *smtpSender) Send(toEmail, toName, subject, body string, attachments []Attachment) error {
	// Compose the email message
	from := mail.Address{Name: s.opts.FromName, Address: s.opts.FromEmail}
	to := mail.Address{Name: toName, Address: toEmail}
	message, err := composeMessage(from, to, subject, body, attachments)
	if err != nil {
		return err
	}

	// Build recipients list
	recipients := []string{toEmail}
	if s.opts.BCC != "" {
		recipients = append(recipients, s.opts.BCC)
	}

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", s.opts.SMTPUsername, s.opts.SMTPPassword, s.opts.SMTPHost)
	err = smtp.SendMail(s.opts.SMTPHost+":"+strconv.Itoa(s.opts.SMTPPort), auth, from.Address, recipients, message)
	if err != nil {
		return err
	}

	return nil
}

// composeMessage builds the raw message for an email with an HTML body.
// Emails with attachments are sent as a multipart message with the HTML body as the first part.
func composeMessage(from, to mail.Address, subject, body string, attachments []Attachment) ([]byte, error) {
	header := "From: " + from.String() + "\r\n" +
		"To: " + to.String() + "\r\n" +
		"Subject: " + subject + "\r\n"

	if len(attachments) == 0 {
		return []byte(header +
			"Content-Type: text/html; charset=utf-8\r\n" +
			"\r\n" +
			body + "\r\n",
		), nil
	}

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	buf.WriteString(header)
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: " + mime.FormatMediaType("multipart/mixed", map[string]string{"boundary": w.Boundary()}) + "\r\n")
	buf.WriteString("\r\n")

	// The body is quoted-printable encoded to keep its lines within the line length limit of SMTP.
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=utf-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return nil, err
	}
	qp := quotedprintable.NewWriter(part)
	if _, err := qp.Write([]byte(body)); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}

	for _, a := range attachments {
		contentType := a.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		part, err := w.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {mime.FormatMediaType(contentType, map[string]string{"name": a.Filename})},
			"Content-Disposition":       {mime.FormatMediaType("attachment", map[string]string{"filename": a.Filename})},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return nil, err
		}

		// Base64 encoded content must be split into lines of at most 76 characters (RFC 2045).
		data := make([]byte, base64.StdEncoding.EncodedLen(len(a.Data)))
		base64.StdEncoding.Encode(data, a.Data)
		for len(data) > 76 {
			if _, err := part.Write(data[:76]); err != nil {
				return nil, err
			}
			if _, err := part.Write([]byte("\r\n")); err != nil {
				return nil, err
			}
			data = data[76:]
		}
		if _, err := part.Write(data); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type consoleSender struct {
	logger    *zap.Logger
	fromEmail string
	fromName  string
}

func NewConsoleSender(logger *zap.Logger, fromEmail, fromName string) (Sender, error) {
	return &consoleSender{logger: logger, fromEmail: fromEmail, fromName: fromName}, nil
}

func (s *consoleSender) Send(toEmail, toName, subject, body string, attachments []Attachment) error {
	attachmentNames := make([]string, len(attachments))
	for i, a := range attachments {
		attachmentNames[i] = a.Filename
	}
	s.logger.Info("email sent",
		zap.String("from_email", s.fromEmail),
		zap.String("from_name", s.fromName),
		zap.String("to_email", toEmail),
		zap.String("to_name", toName),
		zap.String("subject", subject),
		zap.String("body", body),
		zap.Strings("attachments", attachmentNames),
	)
	return nil
}

type noopSender struct{}

func NewNoopSender() Sender {
	return &noopSender{}
}

func (s *noopSender) Send(toEmail, toName, subject, body string, attachments []Attachment) error {
	return nil
}

type TestSender struct {
	Emails []struct {
		ToEmail     string
		ToName      string
		Subject     string
		Body        string
		Attachments []Attachment
	}
}

func NewTestSender() Sender {
	return &TestSender{}
}

func (s *TestSender) Send(toEmail, toName, subject, body string, attachments []Attachment) error {
	s.Emails = append(s.Emails, struct {
		ToEmail     string
		ToName      string
		Subject     string
		Body        string
		Attachments []Attachment
	}{
		ToEmail:     toEmail,
		ToName:      toName,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	})
	return nil
}
