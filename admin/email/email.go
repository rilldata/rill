package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
)

//go:embed templates/gen/*
var templatesFS embed.FS

type Client struct {
	sender      Sender
	frontendURL string
	templates   *template.Template
}

func New(sender Sender, frontendURL string) *Client {
	_, err := url.Parse(frontendURL)
	if err != nil {
		panic(fmt.Errorf("invalid frontendURL: %w", err))
	}

	return &Client{
		sender:      sender,
		frontendURL: frontendURL,
		templates:   template.Must(template.New("").ParseFS(templatesFS, "templates/gen/*.html")),
	}
}

type CallToAction struct {
	ToEmail    string
	ToName     string
	Subject    string
	Title      string
	Body       template.HTML
	ButtonText string
	ButtonLink string
}

func (c *Client) SendCallToAction(opts *CallToAction) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("call_to_action.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

type OrganizationInvite struct {
	ToEmail       string
	ToName        string
	OrgName       string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendOrganizationInvite(opts *OrganizationInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("%s invited you to join Rill", opts.InvitedByName),
		Title:      "Accept your invitation to Rill",
		Body:       template.HTML(fmt.Sprintf("%s has invited you to join <b>%s</b> as a %s for their Rill account. Get started interacting with fast, exploratory dashboards by clicking the button below to sign in and accept your invitation.", opts.InvitedByName, opts.OrgName, opts.RoleName)),
		ButtonText: "Accept invitation",
		ButtonLink: mustJoinURLPath(c.frontendURL, opts.OrgName),
	})
}

type OrganizationAddition struct {
	ToEmail       string
	ToName        string
	OrgName       string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendOrganizationAddition(opts *OrganizationAddition) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("%s has added you to %s", opts.InvitedByName, opts.OrgName),
		Title:      fmt.Sprintf("%s has added you to %s", opts.InvitedByName, opts.OrgName),
		Body:       template.HTML(fmt.Sprintf("%s has added you as a %s for <b>%s</b>. Click the button below to view and collaborate on Rill dashboard projects for %s.", opts.InvitedByName, opts.RoleName, opts.OrgName, opts.OrgName)),
		ButtonText: "View account",
		ButtonLink: mustJoinURLPath(c.frontendURL, opts.OrgName),
	})
}

type ProjectInvite struct {
	ToEmail       string
	ToName        string
	OrgName       string
	ProjectName   string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendProjectInvite(opts *ProjectInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("You have been invited to the %s/%s project", opts.OrgName, opts.ProjectName),
		Title:      fmt.Sprintf("You have been invited to the %s/%s project", opts.OrgName, opts.ProjectName),
		Body:       template.HTML(fmt.Sprintf("%s has invited you to collaborate as a %s for the <b>%s/%s</b> project. Click the button below to accept your invitation. ", opts.InvitedByName, opts.RoleName, opts.OrgName, opts.ProjectName)),
		ButtonText: "Accept invitation",
		ButtonLink: mustJoinURLPath(c.frontendURL, opts.OrgName, opts.ProjectName),
	})
}

type ProjectAddition struct {
	ToEmail       string
	ToName        string
	OrgName       string
	ProjectName   string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendProjectAddition(opts *ProjectAddition) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("You have been added to the %s/%s project", opts.OrgName, opts.ProjectName),
		Title:      fmt.Sprintf("You have been added to the %s/%s project", opts.OrgName, opts.ProjectName),
		Body:       template.HTML(fmt.Sprintf("%s has invited you to collaborate as a %s for the <b>%s</b> project. Click the button below to accept your invitation. ", opts.InvitedByName, opts.RoleName, opts.ProjectName)),
		ButtonText: "View account",
		ButtonLink: mustJoinURLPath(c.frontendURL, opts.OrgName, opts.ProjectName),
	})
}

func mustJoinURLPath(base string, elem ...string) string {
	res, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}
	return res
}
