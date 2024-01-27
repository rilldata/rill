package drivers

import "context"

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
}

type ReportMetadata struct {
	OpenURL   string
	ExportURL string
	EditURL   string
}

type AlertMetadata struct {
	OpenURL            string
	EditURL            string
	QueryForAttributes map[string]any
}
