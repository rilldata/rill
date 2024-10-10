package drivers

import (
	"context"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type AdminService interface {
	GetReportMetadata(ctx context.Context, reportName string, reportSpec *runtimev1.ReportSpec, executionTime time.Time) (*ReportMetadata, error)
	GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*AlertMetadata, error)
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
