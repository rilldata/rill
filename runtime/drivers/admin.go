package drivers

import (
	"context"
	"time"
)

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName, ownerID string, emailRecipients []string, executionTime time.Time) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
	ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error)
}

type ReportMetadata struct {
	BaseURLs      ReportURLs
	RecipientURLs map[string]ReportURLs
}

type ReportURLs struct {
	OpenURL   string
	ExportURL string
	EditURL   string
}

type AlertMetadata struct {
	OpenURL            string
	EditURL            string
	QueryForAttributes map[string]any
}
