package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) TriggerReconcile(ctx context.Context, req *adminv1.TriggerReconcileRequest) (*adminv1.TriggerReconcileResponse, error) {
	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "deployment not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerReconcile(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerReconcileResponse{}, nil
}

func (s *Server) TriggerRefreshSources(ctx context.Context, req *adminv1.TriggerRefreshSourcesRequest) (*adminv1.TriggerRefreshSourcesResponse, error) {
	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "deployment not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerRefreshSources(ctx, depl, req.Sources)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRefreshSourcesResponse{}, nil
}

func (s *Server) TriggerRedeploy(ctx context.Context, req *adminv1.TriggerRedeployRequest) (*adminv1.TriggerRedeployResponse, error) {
	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "deployment not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerRedeploy(ctx, proj, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRedeployResponse{}, nil
}
