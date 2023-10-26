package server

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

func (s *Server) CreateReport(ctx context.Context, req *adminv1.CreateReportRequest) (*adminv1.CreateReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	data, err := yamlForManagedReport(req.Options, claims.OwnerID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	name := uuid.New().String()

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID: proj.ID,
		Branch:    proj.ProdBranch,
		Path:      virtualFilePathForManagedReport(name),
		Data:      data,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitReport(ctx, depl, name)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for report to be created")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile report: %s", err.Error())
	}

	return &adminv1.CreateReportResponse{
		Name: name,
	}, nil
}

func (s *Server) EditReport(ctx context.Context, req *adminv1.EditReportRequest) (*adminv1.EditReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	if spec.Security == nil || !spec.Security.ManagedByUi {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && spec.Security.OwnerUserId == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	data, err := yamlForManagedReport(req.Options, spec.Security.OwnerUserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID: proj.ID,
		Branch:    proj.ProdBranch,
		Path:      virtualFilePathForManagedReport(req.Name),
		Data:      data,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitReport(ctx, depl, req.Name)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for report to be updated")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile report: %s", err.Error())
	}

	return &adminv1.EditReportResponse{}, nil
}

func (s *Server) UnsubscribeReport(ctx context.Context, req *adminv1.UnsubscribeReportRequest) (*adminv1.UnsubscribeReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	if spec.Security == nil || !spec.Security.ManagedByUi {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can unsubscribe from reports")
	}
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	opts, err := recreateReportOptionsFromSpec(spec)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to recreate report options: %s", err.Error())
	}

	found := false
	for idx, email := range opts.Recipients {
		if strings.EqualFold(user.Email, email) {
			opts.Recipients = slices.Delete(opts.Recipients, idx, idx+1)
			found = true
			break
		}
	}

	if !found {
		return nil, status.Error(codes.InvalidArgument, "user is not subscribed to report")
	}

	if len(opts.Recipients) == 0 {
		err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, proj.ProdBranch, virtualFilePathForManagedReport(req.Name))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
		}
	} else {
		data, err := yamlForManagedReport(opts, spec.Security.OwnerUserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
		}

		err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID: proj.ID,
			Branch:    proj.ProdBranch,
			Path:      virtualFilePathForManagedReport(req.Name),
			Data:      data,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update virtual file: %s", err.Error())
		}
	}

	err = s.admin.TriggerReconcileAndAwaitReport(ctx, depl, req.Name)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for report to be updated")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile report: %s", err.Error())
	}

	return &adminv1.UnsubscribeReportResponse{}, nil
}

func (s *Server) DeleteReport(ctx context.Context, req *adminv1.DeleteReportRequest) (*adminv1.DeleteReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	if spec.Security == nil || !spec.Security.ManagedByUi {
		return nil, status.Error(codes.FailedPrecondition, "can't edit report because it was not created from the UI")
	}

	isOwner := claims.OwnerType() == auth.OwnerTypeUser && spec.Security.OwnerUserId == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, proj.ProdBranch, virtualFilePathForManagedReport(req.Name))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete virtual file: %s", err.Error())
	}

	err = s.admin.TriggerReconcileAndAwaitReport(ctx, depl, req.Name)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for report to be deleted")
		}
		return nil, status.Errorf(codes.Internal, "failed to reconcile report: %s", err.Error())
	}

	return &adminv1.DeleteReportResponse{}, nil
}

func (s *Server) TriggerReport(ctx context.Context, req *adminv1.TriggerReportRequest) (*adminv1.TriggerReportResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	isOwner := spec.Security != nil && claims.OwnerType() == auth.OwnerTypeUser && spec.Security.OwnerUserId == claims.OwnerID()
	if !permissions.ManageReports && !isOwner {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit report")
	}

	err = s.admin.TriggerReport(ctx, depl, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to trigger report: %s", err.Error())
	}

	return &adminv1.TriggerReportResponse{}, nil
}

func (s *Server) GenerateReportYAML(ctx context.Context, req *adminv1.GenerateReportYAMLRequest) (*adminv1.GenerateReportYAMLResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	data, err := yamlForCommittedReport(req.Options)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to generate report YAML: %s", err.Error())
	}

	return &adminv1.GenerateReportYAMLResponse{
		Yaml: string(data),
	}, nil
}

func virtualFilePathForManagedReport(name string) string {
	return path.Join("reports", name+".yaml")
}

func yamlForManagedReport(opts *adminv1.ReportOptions, ownerUserID string) ([]byte, error) {
	res := rillv1.ReportYAML{}
	res.Title = opts.Title
	res.Refresh = &rillv1.ScheduleYAML{
		Cron: opts.RefreshCron,
	}
	res.Query.Name = opts.QueryName
	res.Query.ArgsJSON = opts.QueryArgsJson
	res.Export.Format = opts.ExportFormat.String()
	res.Export.Limit = uint(opts.ExportLimit)
	res.Email.Template.OpenURL = opts.OpenUrl
	res.Email.Template.EditURL = ""   // TODO: Add
	res.Email.Template.ExportURL = "" // TODO: Add
	res.Email.Recipients = opts.Recipients
	res.Security.ManagedByUI = true
	res.Security.OwnerUserID = ownerUserID
	return yaml.Marshal(res)
}

func yamlForCommittedReport(opts *adminv1.ReportOptions) ([]byte, error) {
	res := rillv1.ReportYAML{}
	res.Title = opts.Title
	res.Refresh = &rillv1.ScheduleYAML{
		Cron: opts.RefreshCron,
	}
	res.Query.Name = opts.QueryName
	res.Query.ArgsJSON = opts.QueryArgsJson        // TODO: Format as YAML
	res.Export.Format = opts.ExportFormat.String() // TODO: Format as pretty string
	res.Export.Limit = uint(opts.ExportLimit)
	res.Email.Template.OpenURL = opts.OpenUrl
	res.Email.Template.EditURL = ""   // TODO: Add
	res.Email.Template.ExportURL = "" // TODO: Add
	res.Email.Recipients = opts.Recipients
	return yaml.Marshal(res)
}

func recreateReportOptionsFromSpec(spec *runtimev1.ReportSpec) (*adminv1.ReportOptions, error) {
	opts := &adminv1.ReportOptions{}
	opts.Title = spec.Title
	if spec.RefreshSchedule != nil && spec.RefreshSchedule.Cron != "" {
		opts.RefreshCron = spec.RefreshSchedule.Cron
	}
	opts.QueryName = spec.QueryName
	opts.QueryArgsJson = spec.QueryArgsJson
	opts.ExportLimit = spec.ExportLimit
	opts.ExportFormat = spec.ExportFormat
	opts.OpenUrl = spec.EmailOpenUrl
	opts.Recipients = spec.EmailRecipients
	return opts, nil
}
