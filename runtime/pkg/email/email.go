package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
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
	DisplayName    string
	ReportTime     time.Time
	DownloadFormat string
	OpenLink       string
	DownloadLink   string
	EditLink       string
}

type scheduledReportData struct {
	DisplayName      string
	ReportTimeString string // Will be inferred from ReportTime
	DownloadFormat   string
	OpenLink         template.URL
	DownloadLink     template.URL
	EditLink         template.URL
}

func (c *Client) SendScheduledReport(opts *ScheduledReport) error {
	// Build template data
	data := &scheduledReportData{
		DisplayName:      opts.DisplayName,
		ReportTimeString: opts.ReportTime.Format(time.RFC1123),
		DownloadFormat:   opts.DownloadFormat,
		OpenLink:         template.URL(opts.OpenLink),
		DownloadLink:     template.URL(opts.DownloadLink),
		EditLink:         template.URL(opts.EditLink),
	}

	// Build subject
	subject := fmt.Sprintf("%s (%s)", opts.DisplayName, data.ReportTimeString)

	// Resolve template
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("scheduled_report.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

func (c *Client) SendAlertStatus(opts *drivers.AlertStatus) error {
	switch opts.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return c.sendAlertStatus(opts, &alertStatusData{
			DisplayName:         opts.DisplayName,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           opts.IsRecover,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return c.sendAlertFail(opts, &alertFailData{
			DisplayName:         opts.DisplayName,
			ExecutionTimeString: opts.ExecutionTime.Format(time.RFC1123),
			FailRow:             opts.FailRow,
			OpenLink:            template.URL(opts.OpenLink),
			EditLink:            template.URL(opts.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return c.sendAlertStatus(opts, &alertStatusData{
			DisplayName:         opts.DisplayName,
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
	DisplayName         string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertFail(opts *drivers.AlertStatus, data *alertFailData) error {
	subject := fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("alert_fail.html").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()

	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type alertStatusData struct {
	DisplayName         string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	IsPass              bool
	IsRecover           bool
	IsError             bool
	ErrorMessage        string
	OpenLink            template.URL
	EditLink            template.URL
}

func (c *Client) sendAlertStatus(opts *drivers.AlertStatus, data *alertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)
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

type Informational struct {
	ToEmail string
	ToName  string
	Subject string
	Title   string
	Body    template.HTML
}

func (c *Client) SendInformational(opts *Informational) error {
	buf := new(bytes.Buffer)
	err := c.templates.Lookup("informational.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, opts.Subject, html)
}

type OrganizationInvite struct {
	ToEmail       string
	ToName        string
	AcceptURL     string
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
		ButtonLink: opts.AcceptURL,
	})
}

type OrganizationAddition struct {
	ToEmail       string
	ToName        string
	OpenURL       string
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
		ButtonLink: opts.OpenURL,
	})
}

type ProjectInvite struct {
	ToEmail       string
	ToName        string
	AcceptURL     string
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
		ButtonLink: opts.AcceptURL,
	})
}

type ProjectAddition struct {
	ToEmail       string
	ToName        string
	OpenURL       string
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
		ButtonLink: opts.OpenURL,
	})
}

type ProjectAccessRequest struct {
	Title       string
	Body        template.HTML
	ToEmail     string
	ToName      string
	Email       string
	OrgName     string
	ProjectName string
	ApproveLink string
	DenyLink    string
}

func (c *Client) SendProjectAccessRequest(opts *ProjectAccessRequest) error {
	subject := fmt.Sprintf("%s would like to view %s/%s", opts.Email, opts.OrgName, opts.ProjectName)
	if opts.Body == "" {
		opts.Body = template.HTML(fmt.Sprintf("<b>%s</b> would like to view <b>%s/%s</b>", opts.Email, opts.OrgName, opts.ProjectName))
	}

	buf := new(bytes.Buffer)
	err := c.templates.Lookup("project_access_request.html").Execute(buf, opts)
	if err != nil {
		return fmt.Errorf("email template error: %w", err)
	}
	html := buf.String()
	return c.Sender.Send(opts.ToEmail, opts.ToName, subject, html)
}

type ProjectAccessGranted struct {
	ToEmail     string
	ToName      string
	OpenURL     string
	OrgName     string
	ProjectName string
}

func (c *Client) SendProjectAccessGranted(opts *ProjectAccessGranted) error {
	return c.SendCallToAction(&CallToAction{
		ToEmail:    opts.ToEmail,
		ToName:     opts.ToName,
		Subject:    fmt.Sprintf("Your request to %s/%s has been approved", opts.OrgName, opts.ProjectName),
		Title:      "",
		Body:       template.HTML(fmt.Sprintf("Your request to <b>%s/%s</b> has been approved", opts.OrgName, opts.ProjectName)),
		ButtonText: "View project in Rill",
		ButtonLink: opts.OpenURL,
	})
}

type ProjectAccessRejected struct {
	ToEmail     string
	ToName      string
	OrgName     string
	ProjectName string
}

func (c *Client) SendProjectAccessRejected(opts *ProjectAccessRejected) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your request to %s/%s has been denied", opts.OrgName, opts.ProjectName),
		Title:   "",
		Body:    template.HTML(fmt.Sprintf("Your request to <b>%s/%s</b> has been denied. Contact your project admin for help.", opts.OrgName, opts.ProjectName)),
	})
}

type InvoicePaymentFailed struct {
	ToEmail            string
	ToName             string
	OrgName            string
	Currency           string
	Amount             string
	GracePeriodEndDate time.Time
}

func (c *Client) SendInvoicePaymentFailed(opts *InvoicePaymentFailed) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Payment for %s has failed", opts.OrgName),
		Title:   fmt.Sprintf("Payment for %s has failed", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("The payment of %s%s for your %q Rill subscription has failed. Your projects will be hibenrated on %s if payment not received.", opts.Currency, opts.Amount, opts.OrgName, opts.GracePeriodEndDate.Format("January 2, 2006"))),
	})
}

type InvoicePaymentSuccess struct {
	ToEmail  string
	ToName   string
	OrgName  string
	Currency string
	Amount   string
}

// SendInvoicePaymentSuccess Currently Used only when a previously failed invoice payment succeeds
func (c *Client) SendInvoicePaymentSuccess(opts *InvoicePaymentSuccess) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Payment for %s has succeeded", opts.OrgName),
		Title:   fmt.Sprintf("Payment for %s has succeeded", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("The payment of %s%s for your %q Rill subscription has succeeded.", opts.Currency, opts.Amount, opts.OrgName)),
	})
}

type InvoiceUnpaid struct {
	ToEmail string
	ToName  string
	OrgName string
}

// SendInvoiceUnpaid sent after the payment grace period has ended
func (c *Client) SendInvoiceUnpaid(opts *InvoiceUnpaid) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Payment for %s is overdue", opts.OrgName),
		Title:   fmt.Sprintf("Payment for %s is overdue", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("The payment for your Rill subscription on %q is overdue. Your projects have been hibernated.", opts.OrgName)),
	})
}

type SubscriptionEnded struct {
	ToEmail string
	ToName  string
	OrgName string
}

func (c *Client) SendSubscriptionEnded(opts *SubscriptionEnded) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Subscription ended for %s", opts.OrgName),
		Title:   fmt.Sprintf("Subscription ended for %s", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Thank you for using Rill, all your projects have been hibernated as subscription has ended for %q.", opts.OrgName)),
	})
}

type TrialStarted struct {
	ToEmail      string
	ToName       string
	OrgName      string
	TrialEndDate time.Time
}

func (c *Client) SendTrialStarted(opts *TrialStarted) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your trial for %s has started", opts.OrgName),
		Title:   fmt.Sprintf("Your trial for %s has started", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Welcome to Rill! Your trial for %q has started and will end on %s.", opts.OrgName, opts.TrialEndDate.Format("January 2, 2006"))),
	})
}

type TrialEndingSoon struct {
	ToEmail      string
	ToName       string
	OrgName      string
	TrialEndDate time.Time
}

func (c *Client) SendTrialEndingSoon(opts *TrialEndingSoon) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your trial for %s is ending soon", opts.OrgName),
		Title:   fmt.Sprintf("Your trial for %s is ending soon", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Your trial for %q is ending on %s. Upgrade to a paid plan to continue using Rill.", opts.OrgName, opts.TrialEndDate.Format("January 2, 2006"))),
	})
}

type TrialEnded struct {
	ToEmail            string
	ToName             string
	OrgName            string
	GracePeriodEndDate time.Time
}

func (c *Client) SendTrialEnded(opts *TrialEnded) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your trial for %s has ended", opts.OrgName),
		Title:   fmt.Sprintf("Your trial for %s has ended", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Your trial for %q has ended. Your projects will be hibernated on %s. Upgrade to a paid plan to continue using Rill.", opts.OrgName, opts.GracePeriodEndDate.Format("January 2, 2006"))),
	})
}

type TrialGracePeriodEnded struct {
	ToEmail string
	ToName  string
	OrgName string
}

func (c *Client) SendTrialGracePeriodEnded(opts *TrialGracePeriodEnded) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your trial grace period has ended for %s", opts.OrgName),
		Title:   fmt.Sprintf("Your trial grace period has ended for %s", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Your trial grace period has ended for %q. Your projects have been hibernated. Please visit the billing portal to enter payment method and upgrade your plan to continue using Rill.", opts.OrgName)),
	})
}

type TrialExtended struct {
	ToEmail      string
	ToName       string
	OrgName      string
	TrialEndDate time.Time
}

func (c *Client) SendTrialExtended(opts *TrialExtended) error {
	return c.SendInformational(&Informational{
		ToEmail: opts.ToEmail,
		ToName:  opts.ToName,
		Subject: fmt.Sprintf("Your trial for %s has been extended", opts.OrgName),
		Title:   fmt.Sprintf("Your trial for %s has been extened", opts.OrgName),
		Body:    template.HTML(fmt.Sprintf("Welcome to Rill! Your trial for %q has been extended and will end on %s.", opts.OrgName, opts.TrialEndDate.Format("January 2, 2006"))),
	})
}
