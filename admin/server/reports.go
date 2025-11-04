package server

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

func (s *Server) GetReportMeta(ctx context.Context, req *adminv1.GetReportMetaRequest) (*adminv1.GetReportMetaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.report", req.Report),
		attribute.StringSlice("args.email_recipients", req.EmailRecipients),
		attribute.String("args.execution_time", req.ExecutionTime.String()),
		attribute.Bool("args.anon_recipients", req.AnonRecipients),
		attribute.String("args.owner_id", req.OwnerId),
		attribute.String("args.web_open_mode", req.WebOpenMode),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read report meta")
	}

	webOpenMode := WebOpenMode(req.WebOpenMode)
	if webOpenMode == "" {
		webOpenMode = WebOpenModeRecipient // Backwards compatibility during rollout
	}
	if !webOpenMode.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid web open mode %q", req.WebOpenMode)
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	urls := make(map[string]*adminv1.GetReportMetaResponse_URLs)

	var recipients []string
	recipients = append(recipients, req.EmailRecipients...)
	if req.AnonRecipients {
		// add empty email for slack and other notifiers token
		recipients = append(recipients, "")
	}

	var tokens map[string]string
	if webOpenMode == WebOpenModeRecipient {
		// This is the default mode for existing reports, this also implies that reports will break for users who don't have access to the project.
		// But we agree this is acceptable and report owner needs to change to creator mode if they want to share with users who don't have access.
		// In this mode, tokens are used only for unsubscribe links, so no access to resources or owner attributes
		tokens, err = s.createMagicTokens(ctx, proj.OrganizationID, proj.ID, req.Report, "", "", nil, recipients, nil)
	} else {
		// whereFilterJSON and accessibleFields is only needed for backwards compatibility during runtime rollout after admin upgrade, can be removed in next version
		// nolint:staticcheck // needed during rollout for backwards compatibility
		tokens, err = s.createMagicTokens(ctx, proj.OrganizationID, proj.ID, req.Report, req.OwnerId, req.WhereFilterJson, req.AccessibleFields, recipients, req.Resources)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to issue magic auth tokens: %w", err)
	}

	var ownerEmail string
	if req.OwnerId != "" {
		owner, err := s.admin.DB.FindUser(ctx, req.OwnerId)
		if err != nil {
			return nil, err
		}
		ownerEmail = owner.Email
	}

	// Generate URLs for each recipient based on web open mode, and whether they are the owner -
	// 	Owner does not need a token and does not get an unsubscribe link.
	// 	Recipients in creator mode get a token and an unsubscribe link.
	// 	Recipients in recipient mode get an unsubscribe link with token but no token for open/export.
	// 	Recipients in none web open mode do not get an open link.
	// 	Recipients other than owner do not get an edit link, they can edit from the project UI if they have permissions.
	for _, recipient := range recipients {
		if recipient == ownerEmail {
			urls[recipient] = &adminv1.GetReportMetaResponse_URLs{
				OpenUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, tokens[recipient], req.ExecutionTime.AsTime()),
				ExportUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, tokens[recipient]),
				EditUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportEdit(org.Name, proj.Name, req.Report),
			}
			continue
		}
		if webOpenMode == WebOpenModeCreator {
			urls[recipient] = &adminv1.GetReportMetaResponse_URLs{
				OpenUrl:        s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, tokens[recipient], req.ExecutionTime.AsTime()),
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, tokens[recipient]),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient),
			}
		} else if webOpenMode == WebOpenModeRecipient {
			urls[recipient] = &adminv1.GetReportMetaResponse_URLs{
				OpenUrl:        s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, "", req.ExecutionTime.AsTime()),
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, ""),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient), // still use token for unsubscribe so that it works seamlessly for non Rill users
			}
		} else {
			// same as recipient but no open url
			urls[recipient] = &adminv1.GetReportMetaResponse_URLs{
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, ""),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient), // still use token for unsubscribe so that it works seamlessly for non Rill users
			}
		}
	}

	return &adminv1.GetReportMetaResponse{
		RecipientUrls: urls,
	}, nil
}

func (s *Server) CreateReport(ctx context.Context, req *adminv1.CreateReportRequest) (*adminv1.CreateReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.CreateReports {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can create reports")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	name, err := s.generateReportName(ctx, depl, req.Options.DisplayName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	data, err := s.yamlForManagedReport(req.Options, claims.OwnerID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualFilePathForManagedReport(name),
		Data:        data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, name, runtime.ResourceKindReport)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile report: %w", err)
	}

	return &adminv1.CreateReportResponse{
		Name: name,
	}, nil
}

func (s *Server) EditReport(ctx context.Context, req *adminv1.EditReportRequest) (*adminv1.EditReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get report: %s", err.Error())
	}
	annotations := parseReportAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && annotations.AdminOwnerUserID == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	data, err := s.yamlForManagedReport(req.Options, annotations.AdminOwnerUserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualFilePathForManagedReport(req.Name),
		Data:        data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindReport)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile report: %w", err)
	}

	return &adminv1.EditReportResponse{}, nil
}

func (s *Server) UnsubscribeReport(ctx context.Context, req *adminv1.UnsubscribeReportRequest) (*adminv1.UnsubscribeReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get report: %s", err.Error())
	}
	annotations := parseReportAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	if claims.OwnerType() != auth.OwnerTypeUser && claims.OwnerType() != auth.OwnerTypeMagicAuthToken {
		return nil, status.Error(codes.PermissionDenied, "only users can unsubscribe from reports")
	}

	var userEmail string
	var slackEmail string
	if claims.OwnerType() == auth.OwnerTypeUser {
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, err
		}
		userEmail = user.Email
	}

	if claims.OwnerType() == auth.OwnerTypeMagicAuthToken {
		reportTkn, err := s.admin.DB.FindNotificationTokenForMagicAuthToken(ctx, claims.OwnerID())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to find notification token: %s", err.Error())
		}

		if reportTkn.ResourceKind != runtime.ResourceKindReport || reportTkn.ResourceName != req.Name {
			return nil, status.Error(codes.InvalidArgument, "token is not valid for this report")
		}

		if reportTkn.RecipientEmail == "" {
			if req.Email != "" {
				return nil, status.Error(codes.InvalidArgument, "anon token cannot be used for unsubscribing email recipients")
			}
			if req.SlackUser == "" {
				return nil, status.Error(codes.InvalidArgument, "no slack user provided for unsubscribing")
			}
			slackEmail = req.SlackUser
		} else {
			userEmail = reportTkn.RecipientEmail
			if req.Email != "" && !strings.EqualFold(userEmail, req.Email) {
				return nil, status.Error(codes.InvalidArgument, "email does not match token")
			}
		}
	}

	file, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualFilePathForManagedReport(req.Name))
	if err != nil {
		return nil, err
	}

	// Unmarshal file data to reportYAML
	var report reportYAML
	err = yaml.Unmarshal(file.Data, &report)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal report YAML: %s", err.Error())
	}

	found := false
	for idx, email := range report.Notify.Email.Recipients {
		if strings.EqualFold(userEmail, email) {
			report.Notify.Email.Recipients = slices.Delete(report.Notify.Email.Recipients, idx, idx+1)
			found = true
			break
		}
	}
	for idx, email := range report.Notify.Slack.Users {
		if strings.EqualFold(slackEmail, email) {
			report.Notify.Slack.Users = slices.Delete(report.Notify.Slack.Users, idx, idx+1)
			found = true
			break
		}
	}

	if !found {
		return nil, status.Error(codes.InvalidArgument, "user is not subscribed to report")
	}

	if len(report.Notify.Email.Recipients) == 0 && len(report.Notify.Slack.Users) == 0 && len(report.Notify.Slack.Channels) == 0 && len(report.Notify.Slack.Webhooks) == 0 {
		err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, "prod", virtualFilePathForManagedReport(req.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to update virtual file: %w", err)
		}
	} else {
		data, err := yaml.Marshal(report)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
		}

		err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID:   proj.ID,
			Environment: "prod",
			Path:        virtualFilePathForManagedReport(req.Name),
			Data:        data,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update virtual file: %w", err)
		}
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindReport)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile report: %w", err)
	}

	return &adminv1.UnsubscribeReportResponse{}, nil
}

func (s *Server) DeleteReport(ctx context.Context, req *adminv1.DeleteReportRequest) (*adminv1.DeleteReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get report: %s", err.Error())
	}
	annotations := parseReportAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && annotations.AdminOwnerUserID == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, "prod", virtualFilePathForManagedReport(req.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to delete virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, runtime.ResourceKindReport)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile report: %w", err)
	}

	return &adminv1.DeleteReportResponse{}, nil
}

func (s *Server) TriggerReport(ctx context.Context, req *adminv1.TriggerReportRequest) (*adminv1.TriggerReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not get report: %s", err.Error())
	}
	annotations := parseReportAnnotations(spec.Annotations)

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && annotations.AdminOwnerUserID == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	err = s.admin.TriggerReport(ctx, depl, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger report: %w", err)
	}

	return &adminv1.TriggerReportResponse{}, nil
}

func (s *Server) GenerateReportYAML(ctx context.Context, req *adminv1.GenerateReportYAMLRequest) (*adminv1.GenerateReportYAMLResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	data, err := s.yamlForCommittedReport(req.Options)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	return &adminv1.GenerateReportYAMLResponse{
		Yaml: string(data),
	}, nil
}

func (s *Server) yamlForManagedReport(opts *adminv1.ReportOptions, ownerUserID string) ([]byte, error) {
	res := reportYAML{}
	res.Type = "report"
	res.DisplayName = opts.DisplayName
	res.Refresh.Cron = opts.RefreshCron
	res.Refresh.TimeZone = opts.RefreshTimeZone
	res.Watermark = "inherit"
	res.Intervals.Duration = opts.IntervalDuration
	res.Query.Name = opts.QueryName
	res.Query.ArgsJSON = opts.QueryArgsJson
	res.Export.Format = opts.ExportFormat.String()
	res.Export.IncludeHeader = opts.ExportIncludeHeader
	res.Export.Limit = uint(opts.ExportLimit)
	res.Notify.Email.Recipients = opts.EmailRecipients
	res.Notify.Slack.Channels = opts.SlackChannels
	res.Notify.Slack.Users = opts.SlackUsers
	res.Notify.Slack.Webhooks = opts.SlackWebhooks
	res.Annotations.AdminOwnerUserID = ownerUserID
	res.Annotations.AdminManaged = true
	res.Annotations.AdminNonce = time.Now().Format(time.RFC3339Nano)
	res.Annotations.WebOpenPath = opts.WebOpenPath
	res.Annotations.WebOpenState = opts.WebOpenState
	res.Annotations.WebOpenMode = WebOpenMode(opts.WebOpenMode)
	if res.Annotations.WebOpenMode == "" {
		res.Annotations.WebOpenMode = WebOpenModeRecipient // Backwards compatibility
	}
	if !res.Annotations.WebOpenMode.Valid() {
		return nil, fmt.Errorf("invalid web open mode %q", opts.WebOpenMode)
	}
	if opts.Explore != "" && opts.Canvas != "" {
		return nil, fmt.Errorf("cannot set both explore and canvas")
	}
	res.Annotations.Explore = opts.Explore
	res.Annotations.Canvas = opts.Canvas
	return yaml.Marshal(res)
}

func (s *Server) yamlForCommittedReport(opts *adminv1.ReportOptions) ([]byte, error) {
	// Format args as pretty YAML
	var args map[string]interface{}
	if opts.QueryArgsJson != "" {
		err := json.Unmarshal([]byte(opts.QueryArgsJson), &args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse queryArgsJSON: %w", err)
		}
	}

	// Format export format as pretty string
	var exportFormat string
	switch opts.ExportFormat {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		exportFormat = "csv"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		exportFormat = "parquet"
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		exportFormat = "xlsx"
	default:
		exportFormat = opts.ExportFormat.String()
	}

	res := reportYAML{}
	res.Type = "report"
	res.DisplayName = opts.DisplayName
	res.Refresh.Cron = opts.RefreshCron
	res.Refresh.TimeZone = opts.RefreshTimeZone
	res.Watermark = "inherit"
	res.Intervals.Duration = opts.IntervalDuration
	res.Query.Name = opts.QueryName
	res.Query.Args = args
	res.Export.Format = exportFormat
	res.Export.IncludeHeader = opts.ExportIncludeHeader
	res.Export.Limit = uint(opts.ExportLimit)
	res.Notify.Email.Recipients = opts.EmailRecipients
	res.Notify.Slack.Channels = opts.SlackChannels
	res.Notify.Slack.Users = opts.SlackUsers
	res.Notify.Slack.Webhooks = opts.SlackWebhooks
	res.Annotations.WebOpenPath = opts.WebOpenPath
	res.Annotations.WebOpenMode = WebOpenMode(opts.WebOpenMode)
	if res.Annotations.WebOpenMode == "" {
		res.Annotations.WebOpenMode = WebOpenModeRecipient // Backwards compatibility
	}
	if !res.Annotations.WebOpenMode.Valid() {
		return nil, fmt.Errorf("invalid web open mode %q", opts.WebOpenMode)
	}
	return yaml.Marshal(res)
}

// generateReportName generates a random report name with the display name as a seed.
// Example: "My report!" -> "my-report-5b3f7e1a".
// It verifies that the name is not taken (the random component makes any collision unlikely, but we check to be sure).
func (s *Server) generateReportName(ctx context.Context, depl *database.Deployment, displayName string) (string, error) {
	for i := 0; i < 5; i++ {
		name := randomReportName(displayName)

		_, err := s.admin.LookupReport(ctx, depl, name)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				// Success! Name isn't taken
				return name, nil
			}
			return "", fmt.Errorf("failed to check report name: %w", err)
		}
	}

	// Fail-safe in case all names we tried were taken
	return uuid.New().String(), nil
}

func (s *Server) createMagicTokens(ctx context.Context, orgID, projectID, reportName, ownerID, whereFilterJSON string, accessibleFields, emails []string, resources []*adminv1.ResourceName) (map[string]string, error) {
	var createdByUserID *string
	if ownerID != "" {
		createdByUserID = &ownerID
	}
	ttl := 3 * 30 * 24 * time.Hour // approx 3 months
	mgcOpts := &admin.IssueMagicAuthTokenOptions{
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
		FilterJSON:      whereFilterJSON,
		Fields:          accessibleFields,
		Internal:        true,
		TTL:             &ttl,
	}

	var res []database.ResourceName
	for _, r := range resources {
		res = append(res, database.ResourceName{
			Type: r.Type,
			Name: r.Name,
		})
	}

	mgcOpts.Resources = res

	if ownerID != "" {
		// Get the project-level permissions for the creating user.
		orgPerms, err := s.admin.OrganizationPermissionsForUser(ctx, orgID, ownerID)
		if err != nil {
			return nil, err
		}
		projectPermissions, err := s.admin.ProjectPermissionsForUser(ctx, projectID, ownerID, orgPerms)
		if err != nil {
			return nil, err
		}

		// Generate JWT attributes based on the creating user's, but with limited project-level permissions.
		// We store these attributes with the magic token, so it can simulate the creating user (even if the creating user is later deleted or their permissions change).
		//
		// NOTE: A problem with this approach is that if we change the built-in format of JWT attributes, these will remain as they were when captured.
		// NOTE: Another problem is that if the creator is an admin, attrs["admin"] will be true. It shouldn't be a problem today, but could end up leaking some privileges in the future if we're not careful.
		attrs, err := s.jwtAttributesForUser(ctx, ownerID, orgID, projectPermissions)
		if err != nil {
			return nil, err
		}
		mgcOpts.Attributes = attrs
	}

	// issue magic tokens for new external emails
	cctx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	emailTokens := make(map[string]string)
	for _, email := range emails {
		if ownerID == "" {
			// set user attrs as per the email
			mgcOpts.Attributes = map[string]interface{}{
				"name":   "",
				"email":  email,
				"domain": email[strings.LastIndex(email, "@")+1:],
				"groups": []string{},
				"admin":  false,
			}
		}

		tkn, err := s.admin.IssueMagicAuthToken(cctx, mgcOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to issue magic auth token for email %s: %w", email, err)
		}

		emailTokens[email] = tkn.Token().String()

		_, err = s.admin.DB.InsertNotificationToken(cctx, &database.InsertNotificationTokenOptions{
			ResourceKind:     runtime.ResourceKindReport,
			ResourceName:     reportName,
			RecipientEmail:   email,
			MagicAuthTokenID: tkn.Token().ID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert report token for email %s: %w", email, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return emailTokens, nil
}

var reportNameToDashCharsRegexp = regexp.MustCompile(`[ _]+`)

var reportNameExcludeCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func randomReportName(displayName string) string {
	name := reportNameToDashCharsRegexp.ReplaceAllString(displayName, "-")
	name = reportNameExcludeCharsRegexp.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	name = strings.Trim(name, "-")
	if name == "" {
		name = uuid.New().String()
	} else {
		name = name + "-" + uuid.New().String()[0:8]
	}
	return name
}

// reportYAML is derived from runtime/parser.ReportYAML, but adapted for generating (as opposed to parsing) the report YAML.
type reportYAML struct {
	Type        string `yaml:"type"`
	DisplayName string `yaml:"display_name"`
	Title       string `yaml:"title,omitempty"` // Deprecated: replaced by display_name, but kept for backwards compatibility
	Refresh     struct {
		Cron     string `yaml:"cron"`
		TimeZone string `yaml:"time_zone"`
	} `yaml:"refresh"`
	Watermark string `yaml:"watermark"`
	Intervals struct {
		Duration string `yaml:"duration"`
	} `yaml:"intervals"`
	Query struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args,omitempty"`
		ArgsJSON string         `yaml:"args_json,omitempty"`
	} `yaml:"query"`
	Export struct {
		Format        string `yaml:"format"`
		IncludeHeader bool   `yaml:"include_header"`
		Limit         uint   `yaml:"limit"`
	} `yaml:"export"`
	Notify struct {
		Email struct {
			Recipients []string `yaml:"recipients"`
		} `yaml:"email"`
		Slack struct {
			Users    []string `yaml:"users"`
			Channels []string `yaml:"channels"`
			Webhooks []string `yaml:"webhooks"`
		} `yaml:"slack"`
	} `yaml:"notify"`
	Annotations reportAnnotations `yaml:"annotations,omitempty"`
}

type reportAnnotations struct {
	AdminOwnerUserID string      `yaml:"admin_owner_user_id"`
	AdminManaged     bool        `yaml:"admin_managed"`
	AdminNonce       string      `yaml:"admin_nonce"` // To ensure spec version gets updated on writes, to enable polling in TriggerReconcileAndAwaitReport
	WebOpenPath      string      `yaml:"web_open_path"`
	WebOpenState     string      `yaml:"web_open_state"`
	WebOpenMode      WebOpenMode `yaml:"web_open_mode,omitempty"`
	Explore          string      `yaml:"explore,omitempty"`
	Canvas           string      `yaml:"canvas,omitempty"`
}

type WebOpenMode string

const (
	WebOpenModeRecipient WebOpenMode = "recipient"
	WebOpenModeCreator   WebOpenMode = "creator"
	WebOpenModeNone      WebOpenMode = "none"
)

func (m WebOpenMode) Valid() bool {
	switch m {
	case WebOpenModeRecipient, WebOpenModeCreator, WebOpenModeNone:
		return true
	}
	return false
}

func parseReportAnnotations(annotations map[string]string) reportAnnotations {
	if annotations == nil {
		return reportAnnotations{}
	}

	res := reportAnnotations{}
	res.AdminOwnerUserID = annotations["admin_owner_user_id"]
	res.AdminManaged, _ = strconv.ParseBool(annotations["admin_managed"])
	res.AdminNonce = annotations["admin_nonce"]
	res.WebOpenPath = annotations["web_open_path"]
	res.WebOpenState = annotations["web_open_state"]
	res.Explore = annotations["explore"]
	res.Canvas = annotations["canvas"]
	switch annotations["web_open_mode"] {
	case "recipient":
		res.WebOpenMode = WebOpenModeRecipient
	case "creator":
		res.WebOpenMode = WebOpenModeCreator
	case "none":
		res.WebOpenMode = WebOpenModeNone
	case "": // backwards compatibility
		res.WebOpenMode = WebOpenModeRecipient
	}

	return res
}

func virtualFilePathForManagedReport(name string) string {
	return path.Join("reports", name+".yaml")
}
