package admin

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (h *Handle) GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string, executionTime string) (*drivers.ReportMetadata, error) {
	res, err := h.admin.GetReportMeta(ctx, &adminv1.GetReportMetaRequest{
		ProjectId:     h.config.ProjectID,
		Branch:        h.config.Branch,
		Report:        reportName,
		Annotations:   annotations,
		ExecutionTime: executionTime,
	})
	if err != nil {
		return nil, err
	}

	return &drivers.ReportMetadata{
		OpenURL:   res.OpenUrl,
		ExportURL: res.ExportUrl,
		EditURL:   res.EditUrl,
	}, nil
}
