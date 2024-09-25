package admin

import (
	"context"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handle) GetReportMetadata(ctx context.Context, reportName string, reportSpec *runtimev1.ReportSpec, executionTime time.Time) (*drivers.ReportMetadata, error) {
	res, err := h.admin.GetReportMeta(ctx, &adminv1.GetReportMetaRequest{
		ProjectId:     h.config.ProjectID,
		Branch:        h.config.Branch,
		Report:        reportName,
		Spec:          reportSpec,
		ExecutionTime: timestamppb.New(executionTime),
	})
	if err != nil {
		return nil, err
	}

	externalUsersURL := make(map[string]drivers.ReportURLs, len(res.ExternalUsersUrls))
	for k, v := range res.ExternalUsersUrls {
		externalUsersURL[k] = drivers.ReportURLs{
			OpenURL:   v.OpenUrl,
			ExportURL: v.ExportUrl,
			EditURL:   v.EditUrl,
		}
	}

	return &drivers.ReportMetadata{
		InternalUsersURL: drivers.ReportURLs{
			OpenURL:   res.InternalUsersUrls.OpenUrl,
			ExportURL: res.InternalUsersUrls.ExportUrl,
			EditURL:   res.InternalUsersUrls.EditUrl,
		},
		ExternalUsersURL: externalUsersURL,
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
