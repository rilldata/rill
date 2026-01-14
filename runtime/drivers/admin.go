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
	GetDeploymentConfig(ctx context.Context) (*DeploymentConfig, error)
	ListDeployments(ctx context.Context) ([]*Deployment, error)
}

type ReportMetadata struct {
	RecipientURLs map[string]ReportURLs
}

type ReportURLs struct {
	OpenURL        string
	ExportURL      string
	EditURL        string
	UnsubscribeURL string
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

// DeploymentConfig holds configuration returned by the admin service for a deployment.
type DeploymentConfig struct {
	Variables   map[string]string
	Annotations map[string]string
	FrontendURL string
	UpdatedOn   time.Time
	UsesArchive bool
}

type Deployment struct {
	Branch   string
	Editable bool
}
