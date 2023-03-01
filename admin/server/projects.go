package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// FindProjects implements AdminService.
// (GET /v1/organizations/{organization}/projects)
func (s *Server) FindProjects(ctx context.Context, req *adminv1.FindProjectsRequest) (*adminv1.FindProjectsResponse, error) {
	projs, err := s.db.FindProjects(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dtos := make([]*adminv1.Project, len(projs))
	for i, proj := range projs {
		dtos[i] = projToDTO(proj)
	}

	return &adminv1.FindProjectsResponse{Projects: dtos}, nil
}

// (GET /v1/organizations/{organization}/project/{name})
func (s *Server) FindProject(ctx context.Context, req *adminv1.FindProjectRequest) (*adminv1.FindProjectResponse, error) {
	proj, err := s.db.FindProjectByName(ctx, req.Organization, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &adminv1.FindProjectResponse{
		Project: projToDTO(proj),
	}, nil
}

// (POST /v1/organizations/{organization}/projects)
func (s *Server) CreateProject(ctx context.Context, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	org, err := s.db.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	project := &database.Project{
		OrganizationID:     org.ID,
		Name:               req.Name,
		Description:        req.Description,
		GitURL:             sql.NullString{String: req.GitUrl, Valid: true},
		GithubAppInstallID: sql.NullInt64{Int64: req.GithubAppInstallId, Valid: true},
		ProductionBranch:   sql.NullString{String: req.ProductionBranch, Valid: true},
	}
	proj, err := s.db.CreateProject(ctx, org.ID, project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateProjectResponse{
		Project: projToDTO(proj),
	}, nil
}

// (DELETE /v1/organizations/{organization}/project/{name})
func (s *Server) DeleteProject(ctx context.Context, req *adminv1.DeleteProjectRequest) (*adminv1.DeleteProjectResponse, error) {
	proj, err := s.db.FindProjectByName(ctx, req.Organization, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.db.DeleteProject(ctx, proj.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteProjectResponse{
		Name: proj.Name,
	}, nil
}

// (PUT /v1/organizations/{organization}/project/{name})
func (s *Server) UpdateProject(ctx context.Context, req *adminv1.UpdateProjectRequest) (*adminv1.UpdateProjectResponse, error) {
	proj, err := s.db.FindProjectByName(ctx, req.Organization, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj.Description = req.Description
	proj, err = s.db.UpdateProject(ctx, proj)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.UpdateProjectResponse{
		Project: projToDTO(proj),
	}, nil
}

func projToDTO(p *database.Project) *adminv1.Project {
	return &adminv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedOn:   timestamppb.New(p.CreatedOn),
		UpdatedOn:   timestamppb.New(p.CreatedOn),
	}
}
