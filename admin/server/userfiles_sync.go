package server

import (
	"context"

	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SyncProjectFilesToGit enqueues a job that flushes the project's staged user files into its Git repo.
func (s *Server) SyncProjectFilesToGit(ctx context.Context, req *adminv1.SyncProjectFilesToGitRequest) (*adminv1.SyncProjectFilesToGitResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	if !auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to sync project files")
	}

	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project is not connected to a github repository")
	}

	// Clear the outcome of any previous attempt before enqueueing, so the status API reflects this sync.
	err = s.admin.DB.UpdateProjectUserFilesSyncError(ctx, proj.ID, "")
	if err != nil {
		return nil, err
	}

	_, err = s.admin.Jobs.SyncUserFilesToGit(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SyncProjectFilesToGitResponse{}, nil
}

// GetProjectFileSyncStatus reports how many staged user files are pending sync to Git,
// along with the project's auto-sync schedule.
func (s *Server) GetProjectFileSyncStatus(ctx context.Context, req *adminv1.GetProjectFileSyncStatusRequest) (*adminv1.GetProjectFileSyncStatusResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	if !auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	unsynced, err := s.admin.DB.CountStagedVirtualFiles(ctx, proj.ID, "prod")
	if err != nil {
		return nil, err
	}

	res := &adminv1.GetProjectFileSyncStatusResponse{
		UnsyncedCount:       int32(unsynced),
		SyncIntervalSeconds: proj.UserFilesSyncIntervalSeconds,
		LastSyncError:       proj.UserFilesSyncError,
		LastSyncWarning:     proj.UserFilesSyncWarning,
	}
	if proj.UserFilesSyncedOn != nil {
		res.LastSyncedOn = timestamppb.New(*proj.UserFilesSyncedOn)
	}
	return res, nil
}

// UpdateProjectFileSyncSchedule configures periodic auto-sync of the project's user files to Git.
// A zero interval disables auto-sync.
func (s *Server) UpdateProjectFileSyncSchedule(ctx context.Context, req *adminv1.UpdateProjectFileSyncScheduleRequest) (*adminv1.UpdateProjectFileSyncScheduleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.Int64("args.sync_interval_seconds", req.SyncIntervalSeconds),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	if !auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to configure project file sync")
	}

	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project is not connected to a github repository")
	}

	// Enforce a floor so a misconfigured interval can't hammer the repo with commits.
	const minSyncInterval = 600 // 10 minutes, matching the sweep cadence
	if req.SyncIntervalSeconds != 0 && req.SyncIntervalSeconds < minSyncInterval {
		return nil, status.Errorf(codes.InvalidArgument, "sync interval must be 0 (disabled) or at least %d seconds", minSyncInterval)
	}

	err = s.admin.DB.UpdateProjectUserFilesSyncInterval(ctx, proj.ID, req.SyncIntervalSeconds)
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateProjectFileSyncScheduleResponse{
		SyncIntervalSeconds: req.SyncIntervalSeconds,
	}, nil
}
