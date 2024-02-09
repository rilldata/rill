package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

//go:embed templates/gen/*
var templatesFS embed.FS

type Client struct {
	Sender    Sender
	templates *template.Template
}

func New(sender Sender) *Client {
	return &Client{
		Sender:    sender,
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

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type AlertStatus struct {
	ToEmail        string
	ToName         string
	Title          string
	ExecutionTime  time.Time
	Status         runtimev1.AssertionStatus
	IsRecover      bool
	FailRow        map[string]any
	ExecutionError string
	OpenLink       string
	EditLink       string
}

func (c *Client) SendAlertStatus(opts *AlertStatus) error {
	switch opts.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return c.sendAlertStatus(opts, &alertStatusData{
			Title:               opts.Title,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           opts.IsRecover,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return c.sendAlertFail(opts, &alertFailData{
			Title:               opts.Title,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			FailRow:             opts.FailRow,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return c.sendAlertStatus(opts, &alertStatusData{
			Title:               opts.Title,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			IsError:             true,
			ErrorMessage:        opts.ExecutionError,
			OpenLink:            template.URL(opts.EditLink), // NOTE: Using edit link here since for errors, we don't want to open a dashboard, but rather the alert itself
			EditLink:            template.URL(opts.EditLink),
		})
	default:
		return fmt.Errorf("unknown assertion status: %v", opts.Status)
	}
}

type alertFailData struct {
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertFail(opts *AlertStatus, data *alertFailData) error {
	subject := fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert_fail.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type alertStatusData struct {
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	IsPass              bool
	IsRecover           bool
	IsError             bool
	ErrorMessage        string
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertStatus(opts *AlertStatus, data *alertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert_status.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
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
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
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
