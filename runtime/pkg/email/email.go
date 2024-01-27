package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"time"
)

//go:embed templates/gen/*
var templatesFS embed.FS

type Client struct {
	sender    Sender
	templates *template.Template
}

func New(sender Sender) *Client {
	return &Client{
		sender:    sender,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/gen/*.html")),
	}
}

type ScheduledReport struct {
	ToEmail        string
	ToName         string
	Title          string
	ReportTime     time.Time
	DownloadFormat string
	OpenLink       string
	DownloadLink   string
	EditLink       string
}

type scheduledReportData struct {
	Title            string
	ReportTimeString string // Will be inferred from ReportTime
	DownloadFormat   string
	OpenLink         template.URL
	DownloadLink     template.URL
	EditLink         template.URL
}

func (c *Client) SendScheduledReport(opts *ScheduledReport) error {
	// Build template data
	data := &scheduledReportData{
		Title:            opts.Title,
		ReportTimeString: opts.ReportTime.Format(time.RFC1123),
		DownloadFormat:   opts.DownloadFormat,
		OpenLink:         template.URL(opts.OpenLink),
		DownloadLink:     template.URL(opts.DownloadLink),
		EditLink:         template.URL(opts.EditLink),
	}

	// Build subject
	subject := fmt.Sprintf("%s (%s)", opts.Title, data.ReportTimeString)

	// Resolve template
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("scheduled_report.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type Alert struct {
	ToEmail       string
	ToName        string
	Title         string
	ExecutionTime time.Time
	FailRow       map[string]any
	OpenLink      string
	EditLink      string
}

type alertData struct {
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) SendAlert(opts *Alert) error {
	// Build template data
	data := &alertData{
		Title:               opts.Title,
		ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
		FailRow:             opts.FailRow,
		OpenLink:            template.URL(opts.OpenLink),
		EditLink:            template.URL(opts.EditLink),
	}

	// Build subject
	subject := fmt.Sprintf("%s (%s)", opts.Title, data.ExecutionTimeString)

	// Resolve template
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.sender.Send(opts.ToEmail, opts.ToName, subject, html)
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
	AdminURL      string
	FrontendURL   string
	OrgName       string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendOrganizationInvite(opts *OrganizationInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	// Create link URL as "{{ admin URL }}/auth/signup?redirect={{ org frontend URL }}"
	queryParams := url.Values{}
	queryParams.Add("redirect", mustJoinURLPath(opts.FrontendURL, opts.OrgName))
	finalURL := mustJoinURLPath(opts.AdminURL, "/auth/signup") + "?" + queryParams.Encode()

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("%s invited you to join Rill", opts.InvitedByName),
		Title:      "Accept your invitation to Rill",
		Body:       template.HTML(fmt.Sprintf("%s has invited you to join <b>%s</b> as a %s for their Rill account. Get started interacting with fast, exploratory dashboards by clicking the button below to sign in and accept your invitation.", opts.InvitedByName, opts.OrgName, opts.RoleName)),
		ButtonText: "Accept invitation",
		ButtonLink: finalURL,
	})
}

type OrganizationAddition struct {
	ToEmail       string
	ToName        string
	FrontendURL   string
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
		ButtonLink: mustJoinURLPath(opts.FrontendURL, opts.OrgName),
	})
}

type ProjectInvite struct {
	ToEmail       string
	ToName        string
	AdminURL      string
	FrontendURL   string
	OrgName       string
	ProjectName   string
	RoleName      string
	InvitedByName string
}

func (c *Client) SendProjectInvite(opts *ProjectInvite) error {
	if opts.InvitedByName == "" {
		opts.InvitedByName = "Rill"
	}

	// Create link URL as "{{ admin URL }}/auth/signup?redirect={{ project frontend URL }}"
	queryParams := url.Values{}
	queryParams.Add("redirect", mustJoinURLPath(opts.FrontendURL, opts.OrgName, opts.ProjectName))
	finalURL := mustJoinURLPath(opts.AdminURL, "/auth/signup") + "?" + queryParams.Encode()

	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("You have been invited to the %s/%s project", opts.OrgName, opts.ProjectName),
		Title:      fmt.Sprintf("You have been invited to the %s/%s project", opts.OrgName, opts.ProjectName),
		Body:       template.HTML(fmt.Sprintf("%s has invited you to collaborate as a %s for the <b>%s/%s</b> project. Click the button below to accept your invitation. ", opts.InvitedByName, opts.RoleName, opts.OrgName, opts.ProjectName)),
		ButtonText: "Accept invitation",
		ButtonLink: finalURL,
	})
}

type ProjectAddition struct {
	ToEmail       string
	ToName        string
	FrontendURL   string
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
		ButtonLink: mustJoinURLPath(opts.FrontendURL, opts.OrgName, opts.ProjectName),
	})
}

func mustJoinURLPath(base string, elem ...string) string {
	res, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}
	return res
}
