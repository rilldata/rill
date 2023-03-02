package server

import (
	"context"
	"errors"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// FindProjects implements AdminService.
// (GET /v1/organizations/{organization}/projects)
func (s *Server) FindProjects(ctx context.Context, req *adminv1.FindProjectsRequest) (*adminv1.FindProjectsResponse, error) {
	projs, err := s.admin.DB.FindProjects(ctx, req.Organization)
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
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Name)
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
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fullName, err := gitFullName(req.GitUrl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	project := &database.Project{
		OrganizationID:     org.ID,
		Name:               req.Name,
		Description:        req.Description,
		GitURL:             req.GitUrl,
		GitFullName:        fullName,
		GithubAppInstallID: req.GithubAppInstallId,
		ProductionBranch:   req.ProductionBranch,
	}
	proj, err := s.admin.DB.CreateProject(ctx, org.ID, project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateProjectResponse{
		Project: projToDTO(proj),
	}, nil
}

// (DELETE /v1/organizations/{organization}/project/{name})
func (s *Server) DeleteProject(ctx context.Context, req *adminv1.DeleteProjectRequest) (*adminv1.DeleteProjectResponse, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.admin.DB.DeleteProject(ctx, proj.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteProjectResponse{
		Name: proj.Name,
	}, nil
}

// (PUT /v1/organizations/{organization}/project/{name})
func (s *Server) UpdateProject(ctx context.Context, req *adminv1.UpdateProjectRequest) (*adminv1.UpdateProjectResponse, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj.Description = req.Description
	proj.GitURL = req.GitUrl
	proj.GithubAppInstallID = req.GithubAppInstallId

	fullName, err := gitFullName(req.GitUrl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	proj.GitFullName = fullName

	proj, err = s.admin.DB.UpdateProject(ctx, proj)

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

func gitFullName(url string) (string, error) {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	// expected path is /owner/repo.git or /owner/repo
	_, name, _ := strings.Cut(endpoint.Path, "/")
	name, _, _ = strings.Cut(name, ".git")
	return name, nil
}
