package admin

import (
	"context"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	delivery := make(map[string]drivers.ReportDelivery, len(res.DeliveryMeta))
	for k, v := range res.DeliveryMeta {
		resources := make([]*runtimev1.ResourceName, len(v.Resources))
		for i, r := range v.Resources {
			resources[i] = &runtimev1.ResourceName{Kind: r.Type, Name: r.Name}
		}
		delivery[k] = drivers.ReportDelivery{
			OpenURL:           v.OpenUrl,
			ExportURL:         v.ExportUrl,
			EditURL:           v.EditUrl,
			UnsubscribeURL:    v.UnsubscribeUrl,
			UserID:            v.UserId,
			UserAttrs:         v.UserAttrs.AsMap(),
			RestrictResources: v.RestrictResources,
			Resources:         resources,
			EmbedOwner:        v.EmbedOwner,
		}
	}

	return &drivers.ReportMetadata{
		ReportDelivery: delivery,
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

func (h *Handle) GetConfig(ctx context.Context) (*drivers.Config, error) {
	res, err := h.admin.GetDeploymentConfig(ctx, &adminv1.GetDeploymentConfigRequest{
		DeploymentId: "", // Will default to the deployment ID of the current access token.
	})
	if err != nil {
		return nil, err
	}

	return &drivers.Config{
		Variables:             groupVariablesByEnv(res.Variables),
		SystemVariables:       res.SystemVariables,
		Annotations:           res.Annotations,
		FrontendURL:           res.FrontendUrl,
		UpdatedOn:             res.UpdatedOn.AsTime(),
		UsesArchive:           res.UsesArchive,
		DuckdbConnectorConfig: res.DuckdbConnectorConfig.AsMap(),
		Editable:              res.Editable,
	}, nil
}

func (h *Handle) ListDeployments(ctx context.Context) ([]*drivers.Deployment, error) {
	projectResp, err := h.admin.GetProjectByID(ctx, &adminv1.GetProjectByIDRequest{
		Id: h.config.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	resp, err := h.admin.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
		Org:     projectResp.Project.OrgName,
		Project: projectResp.Project.Name,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*drivers.Deployment, 0, len(resp.Deployments))
	for _, d := range resp.Deployments {
		res = append(res, &drivers.Deployment{
			Branch:   d.Branch,
			Editable: d.Editable,
		})
	}

	return res, nil
}

func (h *Handle) SavePersonalCanvas(ctx context.Context, userID, name, displayName, yaml string) (*drivers.PersonalCanvas, error) {
	res, err := h.admin.SavePersonalFileForUser(ctx, &adminv1.SavePersonalFileForUserRequest{
		ProjectId:   h.config.ProjectID,
		UserId:      userID,
		Kind:        runtime.ResourceKindCanvas,
		Name:        name,
		DisplayName: displayName,
		Yaml:        yaml,
	})
	if err != nil {
		return nil, err
	}
	return &drivers.PersonalCanvas{
		Name: res.Name,
		URL:  res.Url,
	}, nil
}

func (h *Handle) ListAlerts(ctx context.Context, userID string) ([]*drivers.ManagedAlert, error) {
	res, err := h.admin.ListAlertsForUser(ctx, &adminv1.ListAlertsForUserRequest{
		ProjectId: h.config.ProjectID,
		UserId:    userID,
	})
	if err != nil {
		return nil, err
	}
	alerts := make([]*drivers.ManagedAlert, 0, len(res.Alerts))
	for _, a := range res.Alerts {
		alerts = append(alerts, &drivers.ManagedAlert{
			Name:    a.Name,
			Version: a.Version,
			Options: alertOptionsFromPB(a.Options),
		})
	}
	return alerts, nil
}

func (h *Handle) SaveAlert(ctx context.Context, userID, name, version string, opts *drivers.AlertOptions) (*drivers.SavedResource, error) {
	pb, err := alertOptionsToPB(opts)
	if err != nil {
		return nil, err
	}
	res, err := h.admin.SaveAlertForUser(ctx, &adminv1.SaveAlertForUserRequest{
		ProjectId: h.config.ProjectID,
		UserId:    userID,
		Name:      name,
		Version:   version,
		Options:   pb,
	})
	if err != nil {
		return nil, err
	}
	return &drivers.SavedResource{
		Name:    res.Name,
		URL:     res.Url,
		Version: res.Version,
	}, nil
}

func (h *Handle) ListReports(ctx context.Context, userID string) ([]*drivers.ManagedReport, error) {
	res, err := h.admin.ListReportsForUser(ctx, &adminv1.ListReportsForUserRequest{
		ProjectId: h.config.ProjectID,
		UserId:    userID,
	})
	if err != nil {
		return nil, err
	}
	reports := make([]*drivers.ManagedReport, 0, len(res.Reports))
	for _, r := range res.Reports {
		reports = append(reports, &drivers.ManagedReport{
			Name:    r.Name,
			Version: r.Version,
			Options: reportOptionsFromPB(r.Options),
		})
	}
	return reports, nil
}

func (h *Handle) SaveReport(ctx context.Context, userID, name, version string, opts *drivers.ReportOptions) (*drivers.SavedResource, error) {
	pb, err := reportOptionsToPB(opts)
	if err != nil {
		return nil, err
	}
	res, err := h.admin.SaveReportForUser(ctx, &adminv1.SaveReportForUserRequest{
		ProjectId: h.config.ProjectID,
		UserId:    userID,
		Name:      name,
		Version:   version,
		Options:   pb,
	})
	if err != nil {
		return nil, err
	}
	return &drivers.SavedResource{
		Name:    res.Name,
		URL:     res.Url,
		Version: res.Version,
	}, nil
}

func alertOptionsToPB(opts *drivers.AlertOptions) (*adminv1.AlertOptions, error) {
	props, err := resolverPropertiesPB(opts.ResolverProperties)
	if err != nil {
		return nil, err
	}
	return &adminv1.AlertOptions{
		DisplayName:          opts.DisplayName,
		RefreshCron:          opts.RefreshCron,
		RefreshTimeZone:      opts.RefreshTimeZone,
		IntervalDuration:     opts.IntervalDuration,
		MetricsViewName:      opts.MetricsViewName,
		QueryName:            opts.QueryName,
		QueryArgsJson:        opts.QueryArgsJSON,
		Resolver:             opts.Resolver,
		ResolverProperties:   props,
		ExportAttach:         opts.ExportAttach,
		ExportFormat:         opts.ExportFormat,
		ExportIncludeHeader:  opts.ExportIncludeHeader,
		ExportLimit:          opts.ExportLimit,
		Renotify:             opts.Renotify,
		RenotifyAfterSeconds: opts.RenotifyAfterSeconds,
		EmailRecipients:      opts.EmailRecipients,
		SlackUsers:           opts.SlackUsers,
		SlackChannels:        opts.SlackChannels,
		SlackWebhooks:        opts.SlackWebhooks,
		WebOpenPath:          opts.WebOpenPath,
		WebOpenState:         opts.WebOpenState,
	}, nil
}

func alertOptionsFromPB(opts *adminv1.AlertOptions) *drivers.AlertOptions {
	return &drivers.AlertOptions{
		DisplayName:          opts.DisplayName,
		RefreshCron:          opts.RefreshCron,
		RefreshTimeZone:      opts.RefreshTimeZone,
		IntervalDuration:     opts.IntervalDuration,
		MetricsViewName:      opts.MetricsViewName,
		QueryName:            opts.QueryName,
		QueryArgsJSON:        opts.QueryArgsJson,
		Resolver:             opts.Resolver,
		ResolverProperties:   resolverPropertiesMap(opts.ResolverProperties),
		ExportAttach:         opts.ExportAttach,
		ExportFormat:         opts.ExportFormat,
		ExportIncludeHeader:  opts.ExportIncludeHeader,
		ExportLimit:          opts.ExportLimit,
		Renotify:             opts.Renotify,
		RenotifyAfterSeconds: opts.RenotifyAfterSeconds,
		EmailRecipients:      opts.EmailRecipients,
		SlackUsers:           opts.SlackUsers,
		SlackChannels:        opts.SlackChannels,
		SlackWebhooks:        opts.SlackWebhooks,
		WebOpenPath:          opts.WebOpenPath,
		WebOpenState:         opts.WebOpenState,
	}
}

func reportOptionsToPB(opts *drivers.ReportOptions) (*adminv1.ReportOptions, error) {
	props, err := resolverPropertiesPB(opts.ResolverProperties)
	if err != nil {
		return nil, err
	}
	return &adminv1.ReportOptions{
		DisplayName:         opts.DisplayName,
		RefreshCron:         opts.RefreshCron,
		RefreshTimeZone:     opts.RefreshTimeZone,
		IntervalDuration:    opts.IntervalDuration,
		QueryName:           opts.QueryName,     // nolint:staticcheck // the report dialog still saves explore reports as queries
		QueryArgsJson:       opts.QueryArgsJSON, // nolint:staticcheck // the report dialog still saves explore reports as queries
		Resolver:            opts.Resolver,
		ResolverProperties:  props,
		ExportFormat:        opts.ExportFormat,
		ExportLimit:         opts.ExportLimit,
		ExportIncludeHeader: opts.ExportIncludeHeader,
		ExportAttach:        opts.ExportAttach,
		EmailRecipients:     opts.EmailRecipients,
		SlackUsers:          opts.SlackUsers,
		SlackChannels:       opts.SlackChannels,
		SlackWebhooks:       opts.SlackWebhooks,
		WebOpenPath:         opts.WebOpenPath,
		WebOpenState:        opts.WebOpenState,
		WebOpenMode:         opts.WebOpenMode,
		Explore:             opts.Explore,
		Canvas:              opts.Canvas,
		MetricsViewFilters:  opts.MetricsViewFilters,
	}, nil
}

func reportOptionsFromPB(opts *adminv1.ReportOptions) *drivers.ReportOptions {
	return &drivers.ReportOptions{
		DisplayName:         opts.DisplayName,
		RefreshCron:         opts.RefreshCron,
		RefreshTimeZone:     opts.RefreshTimeZone,
		IntervalDuration:    opts.IntervalDuration,
		QueryName:           opts.QueryName,     // nolint:staticcheck // the report dialog still saves explore reports as queries
		QueryArgsJSON:       opts.QueryArgsJson, // nolint:staticcheck // the report dialog still saves explore reports as queries
		Resolver:            opts.Resolver,
		ResolverProperties:  resolverPropertiesMap(opts.ResolverProperties),
		ExportFormat:        opts.ExportFormat,
		ExportLimit:         opts.ExportLimit,
		ExportIncludeHeader: opts.ExportIncludeHeader,
		ExportAttach:        opts.ExportAttach,
		EmailRecipients:     opts.EmailRecipients,
		SlackUsers:          opts.SlackUsers,
		SlackChannels:       opts.SlackChannels,
		SlackWebhooks:       opts.SlackWebhooks,
		WebOpenPath:         opts.WebOpenPath,
		WebOpenState:        opts.WebOpenState,
		WebOpenMode:         opts.WebOpenMode,
		Explore:             opts.Explore,
		Canvas:              opts.Canvas,
		MetricsViewFilters:  opts.MetricsViewFilters,
	}
}

// resolverPropertiesPB converts resolver properties to a protobuf struct, keeping nil as nil.
func resolverPropertiesPB(props map[string]any) (*structpb.Struct, error) {
	if props == nil {
		return nil, nil
	}
	return structpb.NewStruct(props)
}

// resolverPropertiesMap converts resolver properties from a protobuf struct, keeping nil as nil.
func resolverPropertiesMap(props *structpb.Struct) map[string]any {
	if props == nil {
		return nil
	}
	return props.AsMap()
}

func (h *Handle) UpdateProjectVariables(ctx context.Context, environment string, variables map[string]string) error {
	projectResp, err := h.admin.GetProjectByID(ctx, &adminv1.GetProjectByIDRequest{
		Id: h.config.ProjectID,
	})
	if err != nil {
		return err
	}

	_, err = h.admin.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
		Org:         projectResp.Project.OrgName,
		Project:     projectResp.Project.Name,
		Environment: environment,
		Variables:   variables,
	})
	return err
}

func groupVariablesByEnv(variables []*adminv1.ProjectVariable) map[string]map[string]string {
	perEnv := make(map[string]map[string]string, len(variables))
	for _, v := range variables {
		vars, ok := perEnv[v.Environment]
		if !ok {
			vars = make(map[string]string)
			perEnv[v.Environment] = vars
		}
		vars[v.Name] = v.Value
	}
	return perEnv
}

// GetConnectorOAuthToken implements drivers.AdminService.
func (h *Handle) GetConnectorOAuthToken(ctx context.Context, connector, configHash, userID string, forceRefresh bool) (*drivers.MCPOAuthToken, error) {
	res, err := h.admin.GetConnectorOAuthToken(ctx, &adminv1.GetConnectorOAuthTokenRequest{
		Connector:    connector,
		ConfigHash:   configHash,
		UserId:       userID,
		ForceRefresh: forceRefresh,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, drivers.ErrNotFound
		}
		return nil, err
	}
	tok := &drivers.MCPOAuthToken{AccessToken: res.AccessToken}
	if res.ExpiresOn != nil {
		tok.ExpiresOn = res.ExpiresOn.AsTime()
	}
	return tok, nil
}

// DeleteConnectorOAuthToken implements drivers.AdminService.
func (h *Handle) DeleteConnectorOAuthToken(ctx context.Context, connector, configHash, userID string) error {
	_, err := h.admin.DeleteConnectorOAuthToken(ctx, &adminv1.DeleteConnectorOAuthTokenRequest{
		Connector:  connector,
		ConfigHash: configHash,
		UserId:     userID,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}
	return nil
}
