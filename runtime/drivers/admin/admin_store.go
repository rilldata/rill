package admin

import (
	"context"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handle) GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	res, err := h.admin.GetReportMeta(ctx, &adminv1.GetReportMetaRequest{
		ProjectId:       h.config.ProjectID,
		Report:          reportName,
		OwnerId:         ownerID,
		EmailRecipients: emailRecipients,
		AnonRecipients:  anonRecipients,
		ExecutionTime:   timestamppb.New(executionTime),
		Resources: []*adminv1.ResourceName{
			{
				Type: runtime.ResourceKindReport,
				Name: reportName,
			},
		},
		WebOpenMode: webOpenMode,
	})
	if err != nil {
		return nil, err
	}

	recipientURLs := make(map[string]drivers.ReportURLs, len(res.RecipientUrls))
	for k, v := range res.RecipientUrls {
		recipientURLs[k] = drivers.ReportURLs{
			OpenURL:        v.OpenUrl,
			ExportURL:      v.ExportUrl,
			EditURL:        v.EditUrl,
			UnsubscribeURL: v.UnsubscribeUrl,
		}
	}

	return &drivers.ReportMetadata{
		RecipientURLs: recipientURLs,
	}, nil
}

func (h *Handle) GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	req := &adminv1.GetAlertMetaRequest{
		ProjectId:       h.config.ProjectID,
		Alert:           alertName,
		Annotations:     annotations,
		OwnerId:         ownerID,
		EmailRecipients: emailRecipients,
		AnonRecipients:  anonRecipients,
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

	recipientURLs := make(map[string]drivers.AlertURLs)
	for email, urls := range res.RecipientUrls {
		recipientURLs[email] = drivers.AlertURLs{
			OpenURL:        urls.OpenUrl,
			EditURL:        urls.EditUrl,
			UnsubscribeURL: urls.UnsubscribeUrl,
		}
	}

	meta := &drivers.AlertMetadata{
		RecipientURLs: recipientURLs,
	}

	if res.QueryForAttributes != nil {
		meta.QueryForAttributes = res.QueryForAttributes.AsMap()
	}

	return meta, nil
}

func (h *Handle) ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error) {
	argsPB, err := structpb.NewStruct(args)
	if err != nil {
		return nil, err
	}

	res, err := h.admin.Provision(ctx, &adminv1.ProvisionRequest{
		DeploymentId: "", // Will default to the deployment ID of the current access token.
		Type:         driver,
		Name:         name,
		Args:         argsPB,
	})
	if err != nil {
		return nil, err
	}

	return res.Resource.Config.AsMap(), nil
}
