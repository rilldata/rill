package drivers

import (
	"context"
	"time"
)

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string, executionTime time.Time) (*ReportMetadata, error)
}

type ReportMetadata struct {
	OpenURL   string
	ExportURL string
	EditURL   string
}
