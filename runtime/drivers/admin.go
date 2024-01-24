package drivers

import "context"

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string, executionTime string) (*ReportMetadata, error)
}

type ReportMetadata struct {
	OpenURL   string
	ExportURL string
	EditURL   string
}
