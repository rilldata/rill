package drivers

import (
	"context"
	"errors"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrNotAuthenticated = errors.New("not authenticated")

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
	ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error)
	GetConfig(ctx context.Context) (*Config, error)
	ListDeployments(ctx context.Context) ([]*Deployment, error)
	UpdateProjectVariables(ctx context.Context, environment string, variables map[string]string) error
	// SavePersonalCanvas creates a personal canvas owned by the given user, or edits it if name is not empty.
	// Annotations in the YAML are ignored: the admin sets the owner and keeps the canvas' sharing state.
	SavePersonalCanvas(ctx context.Context, userID, name, displayName, yaml string) (*PersonalCanvas, error)
	// GetConnectorOAuthToken returns a user's OAuth access token for a connector, refreshing it first if it is about to expire.
	// configHash identifies the connector's current OAuth configuration; a token granted for another configuration is reported as not found.
	// It returns ErrNotFound if the user has not authorized the connector.
	GetConnectorOAuthToken(ctx context.Context, connector, configHash, userID string, forceRefresh bool) (*MCPOAuthToken, error)
	// DeleteConnectorOAuthToken removes a user's OAuth authorizations for a connector, including those granted for its earlier OAuth configurations.
	// configHash identifies the current configuration.
	DeleteConnectorOAuthToken(ctx context.Context, connector, configHash, userID string) error
	// ListAlerts returns the alerts the given user owns, with the options they were saved with.
	ListAlerts(ctx context.Context, userID string) ([]*ManagedAlert, error)
	// SaveAlert creates an alert owned by the given user, or edits one of theirs if name is not empty, as the alert dialog saves it.
	// An edit must pass the version of the alert it's based on, as returned by ListAlerts, and fails if the alert changed since.
	SaveAlert(ctx context.Context, userID, name, version string, opts *AlertOptions) (*SavedResource, error)
	// ListReports returns the reports the given user owns, with the options they were saved with.
	ListReports(ctx context.Context, userID string) ([]*ManagedReport, error)
	// SaveReport creates a report owned by the given user, or edits one of theirs if name is not empty, as the report dialog saves it.
	// An edit must pass the version of the report it's based on, as returned by ListReports, and fails if the report changed since.
	// Unless the report opens with the creator's permissions, it fails for recipients and options that need them, explaining why.
	SaveReport(ctx context.Context, userID, name, version string, opts *ReportOptions) (*SavedResource, error)
}

// PersonalCanvas identifies a saved personal canvas.
type PersonalCanvas struct {
	Name string
	URL  string
}

// MCPOAuthToken is a user's OAuth access token for an external MCP connector.
type MCPOAuthToken struct {
	AccessToken string
	// ExpiresOn is zero if the token does not expire.
	ExpiresOn time.Time
}

// AlertOptions are the options of an alert saved with AdminService.SaveAlert.
// They mirror the admin service's alert options, which the alert dialog sends.
type AlertOptions struct {
	DisplayName          string
	RefreshCron          string
	RefreshTimeZone      string
	IntervalDuration     string
	MetricsViewName      string
	QueryName            string
	QueryArgsJSON        string
	Resolver             string
	ResolverProperties   map[string]any
	ExportAttach         bool
	ExportFormat         runtimev1.ExportFormat
	ExportIncludeHeader  bool
	ExportLimit          uint64
	Renotify             bool
	RenotifyAfterSeconds uint32
	EmailRecipients      []string
	SlackUsers           []string
	SlackChannels        []string
	SlackWebhooks        []string
	WebOpenPath          string
	WebOpenState         string
}

// ReportOptions are the options of a report saved with AdminService.SaveReport.
// They mirror the admin service's report options, which the report dialog sends.
type ReportOptions struct {
	DisplayName         string
	RefreshCron         string
	RefreshTimeZone     string
	IntervalDuration    string
	QueryName           string
	QueryArgsJSON       string
	Resolver            string
	ResolverProperties  map[string]any
	ExportFormat        runtimev1.ExportFormat
	ExportLimit         uint64
	ExportIncludeHeader bool
	ExportAttach        bool
	EmailRecipients     []string
	SlackUsers          []string
	SlackChannels       []string
	SlackWebhooks       []string
	WebOpenPath         string
	WebOpenState        string
	// WebOpenMode is "recipient" (open with each recipient's own permissions), "creator" (open with the creator's, without login) or "none".
	WebOpenMode        string
	Explore            string
	Canvas             string
	MetricsViewFilters map[string]*runtimev1.Expression
}

// SavedResource identifies an alert or report saved through the admin service.
type SavedResource struct {
	Name string
	URL  string
	// Version of the saved alert or report, to edit it again.
	Version string
}

// ManagedAlert is an alert a user owns, as listed by AdminService.ListAlerts.
type ManagedAlert struct {
	Name string
	// Version of the alert, to pass to AdminService.SaveAlert when editing it.
	Version string
	Options *AlertOptions
}

// ManagedReport is a report a user owns, as listed by AdminService.ListReports.
type ManagedReport struct {
	Name string
	// Version of the report, to pass to AdminService.SaveReport when editing it.
	Version string
	Options *ReportOptions
}

type ReportMetadata struct {
	ReportDelivery map[string]ReportDelivery
}

type ReportDelivery struct {
	OpenURL        string
	ExportURL      string
	EditURL        string
	UnsubscribeURL string
	UserID         string         // user ID of the intended recipient, will be empty for non-Rill users and users not having project access. In creator mode this will be the user ID of the creator.
	UserAttrs      map[string]any // user attrs of the intended recipient, will be empty for non-Rill users and users not having project access. In creator mode this will be the user attrs of the creator.
	// EmbedOwner is true if the recipient is the embed user who created the report.
	// Only that recipient gets an open link to the embedder's app.
	EmbedOwner bool
	// RestrictResources is true if the report's data must only be accessed through Resources.
	// It is set for reports created from an embed that was limited to specific resources.
	RestrictResources bool
	Resources         []*runtimev1.ResourceName
}

type AlertURLs struct {
	OpenURL        string
	EditURL        string
	UnsubscribeURL string
}

type AlertMetadata struct {
	RecipientURLs      map[string]AlertURLs
	QueryForAttributes map[string]any
}

// Config holds configuration returned by the admin service for the runtime instance.
// For local runtimes only applicable fields like Variables will be populated.
type Config struct {
	Variables             map[string]map[string]string
	SystemVariables       map[string]string
	Annotations           map[string]string
	FrontendURL           string
	UpdatedOn             time.Time
	UsesArchive           bool
	DuckdbConnectorConfig map[string]any
	Editable              bool
}

type Deployment struct {
	Branch   string
	Editable bool
}
