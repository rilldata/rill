package drivers

import (
	"context"
	"time"
)

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName, ownerID string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
}

type ReportMetadata struct {
	BaseURLs      ReportURLs
	RecipientURLs map[string]ReportURLs
}

type ReportURLs struct {
	OpenURL        string
	ExportURL      string
	UnsubscribeURL string
}

type AlertMetadata struct {
	OpenURL            string
	EditURL            string
	QueryForAttributes map[string]any
}
