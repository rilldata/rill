package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
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
	if !webOpenMode.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid web open mode %q", req.WebOpenMode)
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	delivery := make(map[string]*adminv1.GetReportMetaResponse_DeliveryMeta)

	var recipients []string
	recipients = append(recipients, req.EmailRecipients...)
	if req.AnonRecipients {
		// add empty email for slack and other notifiers token
		recipients = append(recipients, "")
	}

	// Reports created from an embed run as the embed user, and are handled separately.
	embedInfo, err := s.embedCreatedReportInfo(ctx, proj.ID, req.Report)
	if err != nil {
		return nil, err
	}
	if embedInfo != nil {
		return s.getEmbedReportMeta(ctx, org, proj, req, embedInfo, recipients)
	}

	// Resolve the identity to evaluate security policies for when accessing the report's data.
	// NOTE: userID is empty if the report has no owner.
	ownerID := req.OwnerId
	var userID string
	var userAttrs map[string]any
	var ownerEmail string
	if ownerID != "" {
		userAttrs, userID, _, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, ownerID, "")
		if err != nil {
			return nil, err
		}
		owner, err := s.admin.DB.FindUser(ctx, ownerID)
		if err != nil {
			return nil, err
		}
		ownerEmail = owner.Email
	}
	userAttrsPb, err := structpb.NewStruct(userAttrs)
	if err != nil {
		return nil, err
	}

	var tokens map[string]string
	if webOpenMode == WebOpenModeRecipient {
		tokens, err = s.createUnsubMagicTokens(ctx, proj.ID, req.Report, ownerID, ownerEmail, recipients)
	} else {
		tokens, err = s.createMagicTokens(ctx, proj.ID, req.Report, ownerID, recipients, req.Resources, userAttrs)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to issue magic auth tokens: %w", err)
	}

	// Generate URLs for each recipient based on web open mode, and whether they are the owner -
	// 	Owner does not get a token in recipient mode and does not get an unsubscribe link.
	// 	Recipients in creator mode get a token and an unsubscribe link.
	// 	Recipients in recipient mode get an unsubscribe link with token but no token for open/export.
	// 	Recipients in none web open mode do not get an open link.
	// 	Recipients other than owner do not get an edit link, they can edit from the project UI if they have permissions.
	for _, recipient := range recipients {
		if recipient == ownerEmail {
			if webOpenMode == WebOpenModeRecipient && ownerID != "" {
				// owner in recipient mode gets plain open and export url without token as token does not have any access
				delivery[recipient] = &adminv1.GetReportMetaResponse_DeliveryMeta{
					OpenUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, "", req.ExecutionTime.AsTime()),
					ExportUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, ""),
					EditUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportEdit(org.Name, proj.Name, req.Report),
					UserId:    userID,
					UserAttrs: userAttrsPb,
				}
			} else {
				delivery[recipient] = &adminv1.GetReportMetaResponse_DeliveryMeta{
					OpenUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, tokens[recipient], req.ExecutionTime.AsTime()),
					ExportUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, tokens[recipient]),
					EditUrl:   s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportEdit(org.Name, proj.Name, req.Report),
					UserId:    userID,
					UserAttrs: userAttrsPb,
				}
			}
			continue
		}
		if webOpenMode == WebOpenModeCreator {
			delivery[recipient] = &adminv1.GetReportMetaResponse_DeliveryMeta{
				OpenUrl:        s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, tokens[recipient], req.ExecutionTime.AsTime()),
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, tokens[recipient]),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient),
				UserId:         userID,
				UserAttrs:      userAttrsPb,
			}
		} else if webOpenMode == WebOpenModeRecipient {
			attr, userID, err := s.getAttributesForProjectMember(ctx, recipient, proj.OrganizationID, proj.ID)
			if err != nil {
				return nil, err
			}
			pbAttrs, err := structpb.NewStruct(attr)
			if err != nil {
				return nil, err
			}
			delivery[recipient] = &adminv1.GetReportMetaResponse_DeliveryMeta{
				OpenUrl:        s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportOpen(org.Name, proj.Name, req.Report, "", req.ExecutionTime.AsTime()),
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, ""),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient), // still use token for unsubscribe so that it works seamlessly for non Rill users
				UserId:         userID,
				UserAttrs:      pbAttrs,
			}
		} else { // same as recipient but no open url
			attr, userID, err := s.getAttributesForProjectMember(ctx, recipient, proj.OrganizationID, proj.ID)
			if err != nil {
				return nil, err
			}
			pbAttrs, err := structpb.NewStruct(attr)
			if err != nil {
				return nil, err
			}
			delivery[recipient] = &adminv1.GetReportMetaResponse_DeliveryMeta{
				ExportUrl:      s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportExport(org.Name, proj.Name, req.Report, ""),
				UnsubscribeUrl: s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tokens[recipient], recipient), // still use token for unsubscribe so that it works seamlessly for non Rill users
				UserId:         userID,
				UserAttrs:      pbAttrs,
			}
		}
	}

	return &adminv1.GetReportMetaResponse{
		DeliveryMeta: delivery,
	}, nil
}

// embedReportOwnerInactivityLimit is how long reports created from an embed keep running after their owner last opened an iframe.
// Without a limit, the reports of a user who lost access in the embedder's app would keep running indefinitely with the user's last identity.
const embedReportOwnerInactivityLimit = 90 * 24 * time.Hour

// getEmbedReportMeta returns the report metadata for a report created from an embed.
// The report runs as the identity of the embed user's latest iframe, and fails if the embed user is missing or inactive.
// Recipients only get an unsubscribe link, since they open the report through the embedder's app (if at all).
func (s *Server) getEmbedReportMeta(ctx context.Context, org *database.Organization, proj *database.Project, req *adminv1.GetReportMetaRequest, info *embedReportInfo, recipients []string) (*adminv1.GetReportMetaResponse, error) {
	if info.ownerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "the embed user who created the report no longer exists")
	}
	embedUser, err := s.admin.DB.FindEmbedUser(ctx, info.ownerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "the embed user who created the report no longer exists")
		}
		return nil, err
	}
	if time.Since(embedUser.LastSeenOn) > embedReportOwnerInactivityLimit {
		return nil, status.Errorf(codes.FailedPrecondition, "the report is paused because its owner hasn't opened an embedded dashboard in %d days", int(embedReportOwnerInactivityLimit.Hours()/24))
	}

	// Resolve the identity the same way as the embed's runtime token, so the report evaluates security policies identically.
	var email string
	if embedUser.Email != nil {
		email = *embedUser.Email
	}
	user, err := s.resolveEndUser(ctx, proj, "", email, embedUser.Attributes)
	if err != nil {
		return nil, err
	}
	userAttrsPb, err := structpb.NewStruct(mergeExtraAttributes(user.attributes, map[string]any{"embed": true}))
	if err != nil {
		return nil, err
	}

	// Limit the report's data the same way as the embed's runtime token:
	// to the resources the embed was limited to, or else to the resource restrictions of the project member with the embed user's email.
	restrictResources := len(info.resources) > 0
	resources := info.resources
	if !restrictResources {
		restrictResources = user.restrictResources
		resources = user.resources
	}
	resourcesPb := make([]*adminv1.ResourceName, len(resources))
	for i, r := range resources {
		resourcesPb[i] = &adminv1.ResourceName{Type: r.Type, Name: r.Name}
	}

	// Recipients can be any address, so each of them gets an unsubscribe link.
	// The tokens have no resources, so they don't grant access to any data.
	tokens, err := s.createUnsubMagicTokens(ctx, proj.ID, req.Report, "", "", req.EmailRecipients)
	if err != nil {
		return nil, fmt.Errorf("failed to issue magic auth tokens: %w", err)
	}

	delivery := make(map[string]*adminv1.GetReportMetaResponse_DeliveryMeta, len(recipients))
	for _, recipient := range recipients {
		meta := &adminv1.GetReportMetaResponse_DeliveryMeta{
			UserId:            user.userID,
			UserAttrs:         userAttrsPb,
			RestrictResources: restrictResources,
			Resources:         resourcesPb,
		}
		if tkn, ok := tokens[recipient]; ok {
			meta.UnsubscribeUrl = s.admin.URLs.WithCustomDomain(org.CustomDomain).ReportUnsubscribe(org.Name, proj.Name, req.Report, tkn, recipient)
		}
		delivery[recipient] = meta
	}
	return &adminv1.GetReportMetaResponse{
		DeliveryMeta: delivery,
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

	ownerID := ""
	embedOwnerID := ""
	var embedResources []database.ResourceName
	switch claims.OwnerType() {
	case auth.OwnerTypeUser:
		ownerID = claims.OwnerID()
	case auth.OwnerTypeEmbed:
		mdl, ok := claims.AuthTokenModel().(*admin.EmbedToken)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "invalid embed token")
		}
		embedOwnerID = mdl.EmbedUserID
		embedResources, err = embedReportResources(mdl.Resources, req.Options)
		if err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.PermissionDenied, "only users can create reports")
	}

	if proj.PrimaryDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	name, err := s.generateVirtualFileName(ctx, req.Options.DisplayName, func(ctx context.Context, name string) error {
		_, err := s.admin.LookupReport(ctx, depl, name)
		return err
	})
	if err != nil {
		return nil, err
	}

	data, err := s.yamlForManagedReport(req.Options, ownerID, embedOwnerID, embedResources)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	var embedOwnerIDPtr *string
	if embedOwnerID != "" {
		embedOwnerIDPtr = &embedOwnerID
	}
	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:    proj.ID,
		Environment:  "prod",
		Path:         virtualFilePathForManagedReport(name),
		Data:         data,
		OwnerID:      nil,
		EmbedOwnerID: embedOwnerIDPtr,
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

	if proj.PrimaryDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, fmt.Errorf("could not get report: %w", err)
	}
	annotations := parseReportAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}
	embedInfo, err := s.embedCreatedReportInfo(ctx, proj.ID, req.Name)
	if err != nil {
		return nil, err
	}
	if !permissions.ManageReports && !isAdminManagedOwner(claims, annotations.AdminOwnerUserID, embedInfo) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	// Preserve the original owner, since the editor may be someone with manage permissions.
	// If the embed user who created the report no longer exists, editing it would turn it into a report without an owner.
	var embedOwnerID string
	if embedInfo != nil {
		if embedInfo.ownerID == "" {
			return nil, status.Error(codes.FailedPrecondition, "can't edit report because the embed user who created it no longer exists")
		}
		embedOwnerID = embedInfo.ownerID
	}

	// Keep the resources the report was limited to.
	// An embed editor can't give the report more access than its own embed has, so it can only narrow them.
	embedResources, err := parseEmbedResources(annotations.AdminEmbedResources)
	if err != nil {
		return nil, err
	}
	if mdl, ok := claims.AuthTokenModel().(*admin.EmbedToken); ok && len(mdl.Resources) > 0 {
		if len(embedResources) == 0 {
			embedResources = mdl.Resources
		} else {
			embedResources = slices.DeleteFunc(embedResources, func(r database.ResourceName) bool {
				return !slices.Contains(mdl.Resources, r)
			})
			if len(embedResources) == 0 {
				return nil, status.Error(codes.PermissionDenied, "can't edit a report that targets dashboards outside of this embed")
			}
		}
	}
	embedResources, err = embedReportResources(embedResources, req.Options)
	if err != nil {
		return nil, err
	}

	data, err := s.yamlForManagedReport(req.Options, annotations.AdminOwnerUserID, embedOwnerID, embedResources)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	var embedOwnerIDPtr *string
	if embedOwnerID != "" {
		embedOwnerIDPtr = &embedOwnerID
	}
	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:    proj.ID,
		Environment:  "prod",
		Path:         virtualFilePathForManagedReport(req.Name),
		Data:         data,
		OwnerID:      nil,
		EmbedOwnerID: embedOwnerIDPtr,
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
	if req.Email != "" {
		observability.AddRequestAttributes(ctx, attribute.String("args.email", req.Email))
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, fmt.Errorf("could not get report: %w", err)
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
			return nil, fmt.Errorf("failed to find notification token: %w", err)
		}

		if reportTkn.ResourceKind != runtime.ResourceKindReport || reportTkn.ResourceName != req.Name {
			return nil, status.Error(codes.PermissionDenied, "token is not valid for this report")
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
		return nil, status.Error(codes.FailedPrecondition, "user is not subscribed to report")
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
			ProjectID:    proj.ID,
			Environment:  "prod",
			Path:         virtualFilePathForManagedReport(req.Name),
			Data:         data,
			OwnerID:      nil,
			EmbedOwnerID: file.EmbedOwnerID, // Preserve the embed owner of reports created from an embed
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

	if proj.PrimaryDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, fmt.Errorf("could not get report: %w", err)
	}
	annotations := parseReportAnnotations(spec.Annotations)

	if !annotations.AdminManaged {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	embedInfo, err := s.embedCreatedReportInfo(ctx, proj.ID, req.Name)
	if err != nil {
		return nil, err
	}
	if !permissions.ManageReports && !isAdminManagedOwner(claims, annotations.AdminOwnerUserID, embedInfo) {
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

	if proj.PrimaryDeploymentID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	spec, err := s.admin.LookupReport(ctx, depl, req.Name)
	if err != nil {
		return nil, fmt.Errorf("could not get report: %w", err)
	}
	annotations := parseReportAnnotations(spec.Annotations)

	embedInfo, err := s.embedCreatedReportInfo(ctx, proj.ID, req.Name)
	if err != nil {
		return nil, err
	}
	if !permissions.ManageReports && !isAdminManagedOwner(claims, annotations.AdminOwnerUserID, embedInfo) {
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

// yamlForManagedReport generates the YAML for a report created from the UI.
// For reports created from an embed, embedOwnerID is the ID of the embed user who created it,
// and embedResources optionally limits the report's data to the resources the embed was limited to.
func (s *Server) yamlForManagedReport(opts *adminv1.ReportOptions, ownerUserID, embedOwnerID string, embedResources []database.ResourceName) ([]byte, error) {
	res := reportYAML{}
	res.Type = "report"
	res.DisplayName = opts.DisplayName
	res.Refresh.Cron = opts.RefreshCron
	res.Refresh.TimeZone = opts.RefreshTimeZone
	res.Watermark = "inherit"
	res.Intervals.Duration = opts.IntervalDuration

	if opts.ExportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_PDF {
		if opts.Canvas == "" {
			return nil, fmt.Errorf("PDF export requires a canvas")
		}
		if opts.Resolver != "" || opts.QueryName != "" { // nolint:staticcheck // backwards compatibility
			return nil, fmt.Errorf("PDF export does not support a query or data resolver")
		}
	}

	if opts.Resolver != "" {
		res.Data = map[string]any{
			opts.Resolver: opts.ResolverProperties,
		}
	}
	res.Query.Name = opts.QueryName         // nolint:staticcheck // backwards compatibility
	res.Query.ArgsJSON = opts.QueryArgsJson // nolint:staticcheck // backwards compatibility
	res.Export.Format = exportFormatYAML(opts.ExportFormat)
	res.Export.IncludeHeader = opts.ExportIncludeHeader
	res.Export.Limit = uint(opts.ExportLimit)
	res.Export.Attach = opts.ExportAttach

	res.Notify.Email.Recipients = opts.EmailRecipients
	res.Notify.Slack.Channels = opts.SlackChannels
	res.Notify.Slack.Users = opts.SlackUsers
	res.Notify.Slack.Webhooks = opts.SlackWebhooks
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

	// Rill users are identified by the owner annotation, which the report reconciler resolves to the user's attributes.
	// Embed users are identified by the embed owner annotation; the admin service resolves it to the identity of the embed user's latest iframe.
	if embedOwnerID != "" {
		// TODO: Support other web open modes for embed users. They currently require recipients to be Rill project members.
		if res.Annotations.WebOpenMode != WebOpenModeCreator {
			return nil, fmt.Errorf("reports created from an embed only support the %q web open mode", WebOpenModeCreator)
		}
		// Reports created from an embed are only delivered by email.
		if len(opts.SlackChannels) > 0 || len(opts.SlackUsers) > 0 || len(opts.SlackWebhooks) > 0 {
			return nil, fmt.Errorf("reports created from an embed can't be sent to Slack")
		}
		// Embed users can only export metrics queries, which are subject to the same security policies as the embedded dashboard.
		// Other resolvers, such as "sql" or "ai", could access data that the embed doesn't expose.
		if opts.Resolver != "" {
			return nil, fmt.Errorf("reports created from an embed must use a query, not a data resolver")
		}
		res.Annotations.AdminEmbedOwnerID = embedOwnerID
		if len(embedResources) > 0 {
			resourcesJSON, err := json.Marshal(embedResources)
			if err != nil {
				return nil, err
			}
			res.Annotations.AdminEmbedResources = string(resourcesJSON)
		}
		// The emails of reports created from an embed have no download link, so they must always attach the export.
		res.Export.Attach = true
	} else {
		res.Annotations.AdminOwnerUserID = ownerUserID
	}
	if res.Export.Attach {
		if opts.QueryName == "" { // nolint:staticcheck // backwards compatibility
			return nil, fmt.Errorf("export attachments are only supported for query-based reports")
		}
		if res.Annotations.WebOpenMode != WebOpenModeCreator {
			return nil, fmt.Errorf("export attachments are only supported for reports run as the creator")
		}
	}
	if opts.Explore != "" && opts.Canvas != "" {
		return nil, fmt.Errorf("cannot set both explore and canvas")
	}
	res.Annotations.Explore = opts.Explore
	res.Annotations.Canvas = opts.Canvas
	if opts.Canvas != "" {
		filtersJSON, err := metricsViewFiltersJSON(opts.MetricsViewFilters)
		if err != nil {
			return nil, err
		}
		res.Annotations.MetricsViewFilters = filtersJSON
	}
	return yaml.Marshal(res)
}

func (s *Server) yamlForCommittedReport(opts *adminv1.ReportOptions) ([]byte, error) {
	// Format args as pretty YAML
	var args map[string]interface{}
	if opts.QueryArgsJson != "" { // nolint:staticcheck // backwards compatibility
		err := json.Unmarshal([]byte(opts.QueryArgsJson), &args) // nolint:staticcheck // backwards compatibility
		if err != nil {
			return nil, fmt.Errorf("failed to parse queryArgsJSON: %w", err)
		}
	}

	res := reportYAML{}
	res.Type = "report"
	res.DisplayName = opts.DisplayName
	res.Refresh.Cron = opts.RefreshCron
	res.Refresh.TimeZone = opts.RefreshTimeZone
	res.Watermark = "inherit"
	res.Intervals.Duration = opts.IntervalDuration
	if opts.Resolver != "" {
		res.Data = map[string]any{
			opts.Resolver: opts.ResolverProperties,
		}
	}
	if opts.ExportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_PDF {
		if opts.Canvas == "" {
			return nil, fmt.Errorf("PDF export requires a canvas")
		}
		if opts.Resolver != "" || opts.QueryName != "" { // nolint:staticcheck // backwards compatibility
			return nil, fmt.Errorf("PDF export does not support a query or data resolver")
		}
	}

	res.Query.Name = opts.QueryName // nolint:staticcheck // backwards compatibility
	res.Query.Args = args
	res.Export.Format = exportFormatYAML(opts.ExportFormat)
	res.Export.IncludeHeader = opts.ExportIncludeHeader
	res.Export.Limit = uint(opts.ExportLimit)
	res.Export.Attach = opts.ExportAttach
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
	if opts.ExportAttach {
		if opts.QueryName == "" { // nolint:staticcheck // backwards compatibility
			return nil, fmt.Errorf("export attachments are only supported for query-based reports")
		}
		if res.Annotations.WebOpenMode != WebOpenModeCreator {
			return nil, fmt.Errorf("export attachments are only supported for reports run as the creator")
		}
	}
	if opts.Explore != "" && opts.Canvas != "" {
		return nil, fmt.Errorf("cannot set both explore and canvas")
	}
	res.Annotations.Explore = opts.Explore
	res.Annotations.Canvas = opts.Canvas
	if opts.Canvas != "" {
		// Canvas reports need the captured canvas state and filters to render correctly.
		res.Annotations.WebOpenState = opts.WebOpenState
		filtersJSON, err := metricsViewFiltersJSON(opts.MetricsViewFilters)
		if err != nil {
			return nil, err
		}
		res.Annotations.MetricsViewFilters = filtersJSON
	}
	return yaml.Marshal(res)
}

// reportFilterMaxSize limits the total size of the serialized filter expressions stored in a report's annotations.
// Mirrors magicAuthTokenFilterMaxSize for magic auth token filters.
const reportFilterMaxSize = 1024

// metricsViewFiltersJSON serializes per-metrics-view filter expressions to a JSON object
// of metrics view name to filter expression in protojson format.
// It is stored in the "metrics_view_filters" report annotation and parsed by the report reconciler.
func metricsViewFiltersJSON(filters map[string]*runtimev1.Expression) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}
	var filterSize int
	res := make(map[string]json.RawMessage, len(filters))
	for mv, expr := range filters {
		if mv == "" {
			return "", fmt.Errorf("empty metrics view name in metrics view filters")
		}
		data, err := protojson.Marshal(expr)
		if err != nil {
			return "", fmt.Errorf("failed to serialize filter for metrics view %q: %w", mv, err)
		}
		filterSize += len(data)
		if filterSize > reportFilterMaxSize {
			return "", fmt.Errorf("filter size exceeds limit of %d bytes", reportFilterMaxSize)
		}
		res[mv] = data
	}
	data, err := json.Marshal(res)
	if err != nil {
		return "", fmt.Errorf("failed to serialize metrics view filters: %w", err)
	}
	return string(data), nil
}

func (s *Server) createMagicTokens(ctx context.Context, projectID, reportName, ownerID string, emails []string, resources []*adminv1.ResourceName, userAttrs map[string]any) (map[string]string, error) {
	var createdByUserID *string
	if ownerID != "" {
		createdByUserID = &ownerID
	}
	ttl := 3 * 30 * 24 * time.Hour // approx 3 months
	mgcOpts := &admin.IssueMagicAuthTokenOptions{
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
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

	// We store the attributes with the magic token, so it can simulate the user (even if the user is later deleted or their permissions change).
	//
	// NOTE: A problem with this approach is that if we change the built-in format of JWT attributes, these will remain as they were when captured.
	// NOTE: Another problem is that if the user is an admin, attrs["admin"] will be true. It shouldn't be a problem today, but could end up leaking some privileges in the future if we're not careful.
	mgcOpts.Attributes = userAttrs

	// issue magic tokens
	cctx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	emailTokens := make(map[string]string)
	for _, email := range emails {
		if userAttrs == nil {
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

func (s *Server) createUnsubMagicTokens(ctx context.Context, projectID, reportName, ownerID, ownerEmail string, emails []string) (map[string]string, error) {
	var createdByUserID *string
	if ownerID != "" {
		createdByUserID = &ownerID
	}
	ttl := 3 * 30 * 24 * time.Hour // approx 3 months
	mgcOpts := &admin.IssueMagicAuthTokenOptions{
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
		Internal:        true,
		TTL:             &ttl,
	}

	// issue magic tokens
	cctx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	emailTokens := make(map[string]string)
	for _, email := range emails {
		if ownerEmail != "" && strings.EqualFold(ownerEmail, email) {
			// skip creating unsubscribe token for owner email
			continue
		}
		// set user attrs as per the email
		mgcOpts.Attributes = map[string]interface{}{
			"name":   "",
			"email":  email,
			"domain": email[strings.LastIndex(email, "@")+1:],
			"groups": []string{},
			"admin":  false,
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

// embedReportInfo describes an admin-managed report that was created from an embed.
type embedReportInfo struct {
	// ownerID is the ID of the embed user who created the report.
	// It is empty if the embed user no longer exists.
	ownerID string
	// resources the report is limited to; an empty list means the report is not limited.
	resources []database.ResourceName
}

// embedCreatedReportInfo returns info about an admin-managed report that was created from an embed, or nil if it wasn't created from an embed.
// The owner is read from the virtual file's embed owner column, which is the source of truth for ownership;
// the report's annotation is only used to detect reports whose embed owner was deleted.
func (s *Server) embedCreatedReportInfo(ctx context.Context, projectID, report string) (*embedReportInfo, error) {
	path := virtualFilePathForManagedReport(report)
	vf, err := s.admin.DB.FindVirtualFile(ctx, projectID, "prod", path)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if vf.Deleted {
		return nil, nil
	}

	var res struct {
		Annotations struct {
			AdminEmbedOwnerID   string `yaml:"admin_embed_owner_id"`
			AdminEmbedResources string `yaml:"admin_embed_resources"`
		} `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(vf.Data, &res); err != nil {
		return nil, fmt.Errorf("failed to parse virtual file %q: %w", path, err)
	}
	if vf.EmbedOwnerID == nil && res.Annotations.AdminEmbedOwnerID == "" {
		return nil, nil
	}

	resources, err := parseEmbedResources(res.Annotations.AdminEmbedResources)
	if err != nil {
		return nil, fmt.Errorf("failed to parse virtual file %q: %w", path, err)
	}
	info := &embedReportInfo{resources: resources}
	if vf.EmbedOwnerID != nil {
		info.ownerID = *vf.EmbedOwnerID
	}
	return info, nil
}

// embedReportResources checks that a report targets a dashboard the embed is limited to,
// and returns the resources to limit the report's data to.
// An empty list of resources means the embed is not limited, in which case it returns an empty list as well.
// Reports and alerts are dropped from the returned resources, since their transitive access is derived from their own queries,
// which can be edited by the embed user. For example, an embed that opens a report is limited to the report and its dashboard.
func embedReportResources(resources []database.ResourceName, opts *adminv1.ReportOptions) ([]database.ResourceName, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	var target database.ResourceName
	switch {
	case opts.Explore != "":
		target = database.ResourceName{Type: runtime.ResourceKindExplore, Name: opts.Explore}
	case opts.Canvas != "":
		target = database.ResourceName{Type: runtime.ResourceKindCanvas, Name: opts.Canvas}
	default:
		return nil, status.Error(codes.PermissionDenied, "reports created from this embed must target an explore or canvas dashboard")
	}
	if !slices.Contains(resources, target) {
		return nil, status.Errorf(codes.PermissionDenied, "reports created from this embed can't target %s %q", target.Type, target.Name)
	}

	var res []database.ResourceName
	for _, r := range resources {
		if r.Type == runtime.ResourceKindExplore || r.Type == runtime.ResourceKindCanvas {
			res = append(res, r)
		}
	}
	return res, nil
}

// parseEmbedResources parses the "admin_embed_resources" annotation of a report created from an embed.
func parseEmbedResources(annotation string) ([]database.ResourceName, error) {
	if annotation == "" {
		return nil, nil
	}
	var res []database.ResourceName
	if err := json.Unmarshal([]byte(annotation), &res); err != nil {
		return nil, fmt.Errorf("invalid admin_embed_resources annotation: %w", err)
	}
	return res, nil
}

// isAdminManagedOwner returns true if the claims belong to the owner of an admin-managed report.
// Rill users are matched by user ID; embed users are matched by embed user ID against the report's embed info (nil if not created from an embed).
func isAdminManagedOwner(claims auth.Claims, ownerUserID string, embedInfo *embedReportInfo) bool {
	switch claims.OwnerType() {
	case auth.OwnerTypeUser:
		return ownerUserID != "" && ownerUserID == claims.OwnerID()
	case auth.OwnerTypeEmbed:
		tkn, ok := claims.AuthTokenModel().(*admin.EmbedToken)
		return ok && embedInfo != nil && embedInfo.ownerID != "" && embedInfo.ownerID == tkn.EmbedUserID
	default:
		return false
	}
}

// getAttributesForProjectMember returns user attributes for a given email only if the user is a member of the project.
func (s *Server) getAttributesForProjectMember(ctx context.Context, email, orgID, projID string) (map[string]any, string, error) {
	if email == "" {
		return nil, "", nil
	}
	// Look up user by email
	user, err := s.admin.DB.FindUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, "", err
	}
	if user == nil {
		return nil, "", nil
	}

	attr, id, readProd, err := s.getAttributesForUser(ctx, orgID, projID, user.ID, "")
	if err != nil {
		return nil, "", err
	}
	if !readProd {
		return nil, "", nil
	}
	return attr, id, nil
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
	Data  map[string]any `yaml:"data,omitempty"` // Generic data resolver block (e.g., data.ai, data.sql)
	Query struct {       // Legacy query-based report (deprecated - use data instead)
		Name     string         `yaml:"name,omitempty"`
		Args     map[string]any `yaml:"args,omitempty"`
		ArgsJSON string         `yaml:"args_json,omitempty"`
	} `yaml:"query,omitempty"`
	Export struct {
		Format        string `yaml:"format"`
		IncludeHeader bool   `yaml:"include_header"`
		Limit         uint   `yaml:"limit"`
		Attach        bool   `yaml:"attach,omitempty"` // Omitted when false, so runtimes that predate it can parse the YAML
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
	AdminOwnerUserID    string      `yaml:"admin_owner_user_id"`
	AdminEmbedOwnerID   string      `yaml:"admin_embed_owner_id,omitempty"`  // ID of the embed user who created the report; the admin service resolves it to the user's identity
	AdminEmbedResources string      `yaml:"admin_embed_resources,omitempty"` // JSON list of resources that limit the data of a report created from an embed
	AdminManaged        bool        `yaml:"admin_managed"`
	AdminNonce          string      `yaml:"admin_nonce"` // To ensure spec version gets updated on writes, to enable polling in TriggerReconcileAndAwaitReport
	WebOpenPath         string      `yaml:"web_open_path"`
	WebOpenState        string      `yaml:"web_open_state"`
	WebOpenMode         WebOpenMode `yaml:"web_open_mode,omitempty"`
	Explore             string      `yaml:"explore,omitempty"`
	Canvas              string      `yaml:"canvas,omitempty"`
	// Per-metrics-view filters of the canvas at scheduling time, as a JSON object of
	// metrics view name to filter expression in protojson format (canvas reports only).
	// The report reconciler bakes these into the report's transitive security rules.
	MetricsViewFilters string `yaml:"metrics_view_filters,omitempty"`
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
	res.AdminEmbedOwnerID = annotations["admin_embed_owner_id"]
	res.AdminEmbedResources = annotations["admin_embed_resources"]
	res.AdminManaged, _ = strconv.ParseBool(annotations["admin_managed"])
	res.AdminNonce = annotations["admin_nonce"]
	res.WebOpenPath = annotations["web_open_path"]
	res.WebOpenState = annotations["web_open_state"]
	res.Explore = annotations["explore"]
	res.Canvas = annotations["canvas"]
	res.MetricsViewFilters = annotations["metrics_view_filters"]
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

// exportFormatYAML formats an export format enum as the string used in report YAML (parsed by runtime/parser.parseExportFormat).
func exportFormatYAML(f runtimev1.ExportFormat) string {
	switch f {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return "csv"
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return "xlsx"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return "parquet"
	case runtimev1.ExportFormat_EXPORT_FORMAT_PDF:
		return "pdf"
	default:
		return f.String()
	}
}

func virtualFilePathForManagedReport(name string) string {
	return path.Join("reports", name+".yaml")
}
