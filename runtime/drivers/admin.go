package drivers

import "context"

type AdminStore interface {
	GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string) (*ReportMetadata, error)
}

type ReportMetadata struct {
	OpenURL   string
	ExportURL string
	EditURL   string
}
