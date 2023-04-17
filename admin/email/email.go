package email

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"
)

type Email interface {
	SendOrganizationInvite(toEmail, toName, orgName, roleName string) error
	SendProjectInvite(toEmail, toName, projectName, roleName string) error
}

type client struct {
	opts *Options
}

type Options struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SenderEmail  string
	SenderName   string
	FrontendURL  string
}

func NewEmail(opts *Options) (Email, error) {
	if opts.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP server password is required")
	}
	_, err := mail.ParseAddress(opts.SenderEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid sender email address %q", opts.SenderEmail)
	}

	return &client{opts: opts}, nil
}

func (c *client) SendOrganizationInvite(toEmail, toName, orgName, roleName string) error {
	err := c.sendMail(toEmail, toName, "Invitation to join Rill",
		fmt.Sprintf("You have been invited to organization <b>%s</b> as <b>%s</b>. Please sign into Rill <a href=\"%s\">here</a> to accept invitation.", orgName, roleName, c.opts.FrontendURL))
	return err
}

func (c *client) SendProjectInvite(toEmail, toName, projectName, roleName string) error {
	err := c.sendMail(toEmail, toName, "Invitation to join Rill",
		fmt.Sprintf("You have been invited to project <b>%s</b> as <b>%s</b>. Please sign into Rill <a href=\"%s\">here</a> to accept invitation.", projectName, roleName, c.opts.FrontendURL))
	return err
}

func (c *client) sendMail(toEmail, toName, subject, body string) error {
	from := mail.Address{Name: c.opts.SenderName, Address: c.opts.SenderEmail}
	to := mail.Address{Name: toName, Address: toEmail}

	// Compose the email message
	message := []byte("From: " + from.String() + "\r\n" +
		"To: " + to.String() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", c.opts.SMTPUsername, c.opts.SMTPPassword, c.opts.SMTPHost)
	err := smtp.SendMail(c.opts.SMTPHost+":"+strconv.Itoa(c.opts.SMTPPort), auth, from.Address, []string{to.Address}, message)
	if err != nil {
		return err
	}
	return nil
}
