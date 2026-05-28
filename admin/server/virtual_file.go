package server

import (
	"context"
	"fmt"
	"path"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetPersonalFile(ctx context.Context, req *adminv1.GetPersonalFileRequest) (*adminv1.GetPersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualFilePathForPersonalFile(req.Name, auth.GetClaims(ctx).OwnerID()))
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	return &adminv1.GetPersonalFileResponse{
		Yaml: string(vf.Data),
	}, nil
}

func (s *Server) PutPersonalFile(ctx context.Context, req *adminv1.PutPersonalFileRequest) (*adminv1.PutPersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualFilePathForPersonalFile(req.Name, auth.GetClaims(ctx).OwnerID()),
		Data:        []byte(req.Yaml),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create personal file: %w", err)
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	err = s.admin.TriggerParser(ctx, depl)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger parser: %w", err)
	}

	return nil, nil
}

func virtualFilePathForPersonalFile(name, userID string) string {
	return path.Join("personal", fmt.Sprintf("%s_%s.yaml", name, userID))
}
