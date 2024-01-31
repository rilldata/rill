package admin

import (
	"context"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handle) GetReportMetadata(ctx context.Context, reportName string, annotations map[string]string, executionTime time.Time) (*drivers.ReportMetadata, error) {
	res, err := h.admin.GetReportMeta(ctx, &adminv1.GetReportMetaRequest{
		ProjectId:     h.config.ProjectID,
		Branch:        h.config.Branch,
		Report:        reportName,
		Annotations:   annotations,
		ExecutionTime: timestamppb.New(executionTime),
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

func (h *Handle) GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	req := &adminv1.GetAlertMetaRequest{
		ProjectId:   h.config.ProjectID,
		Branch:      h.config.Branch,
		Alert:       alertName,
		Annotations: annotations,
	}

	if queryForUserID != "" {
		req.QueryFor = &adminv1.GetAlertMetaRequest_QueryForUserId{QueryForUserId: queryForUserID}
	} else if queryForUserEmail != "" {
		req.QueryFor = &adminv1.GetAlertMetaRequest_QueryForUserEmail{QueryForUserEmail: queryForUserEmail}
	}

	res, err := h.admin.GetAlertMeta(ctx, req)
	if err != nil {
		return nil, err
	}

	meta := &drivers.AlertMetadata{
		OpenURL: res.OpenUrl,
		EditURL: res.EditUrl,
	}

	if res.QueryForAttributes != nil {
		meta.QueryForAttributes = res.QueryForAttributes.AsMap()
	}

	return meta, nil
}
