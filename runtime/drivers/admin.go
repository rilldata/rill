package drivers

import (
	"context"
	"errors"
	"time"
)

var ErrNotAuthenticated = errors.New("not authenticated")

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
	ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error)
	GetConfig(ctx context.Context) (*Config, error)
	ListDeployments(ctx context.Context) ([]*Deployment, error)
	UpdateProjectVariables(ctx context.Context, environment string, variables map[string]string) error
	// GetConnectorOAuthToken returns a user's OAuth access token for a connector, refreshing it first if it is about to expire.
	// configHash identifies the connector's current OAuth configuration; a token granted for another configuration is reported as not found.
	// It returns ErrNotFound if the user has not authorized the connector.
	GetConnectorOAuthToken(ctx context.Context, connector, configHash, userID string, forceRefresh bool) (*MCPOAuthToken, error)
	// DeleteConnectorOAuthToken removes a user's OAuth authorization for a connector's OAuth configuration.
	DeleteConnectorOAuthToken(ctx context.Context, connector, configHash, userID string) error
}

// MCPOAuthToken is a user's OAuth access token for an external MCP connector.
type MCPOAuthToken struct {
	AccessToken string
	// ExpiresOn is zero if the token does not expire.
	ExpiresOn time.Time
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
