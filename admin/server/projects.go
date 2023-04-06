package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListProjects(ctx context.Context, req *adminv1.ListProjectsRequest) (*adminv1.ListProjectsResponse, error) {
	projs, err := s.admin.DB.FindProjects(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dtos := make([]*adminv1.Project, len(projs))
	for i, proj := range projs {
		dtos[i] = projToDTO(proj)
	}

	return &adminv1.ListProjectsResponse{Projects: dtos}, nil
}

func (s *Server) GetProject(ctx context.Context, req *adminv1.GetProjectRequest) (*adminv1.GetProjectResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if proj.ProductionDeploymentID == nil {
		return &adminv1.GetProjectResponse{
			Project: projToDTO(proj),
		}, nil
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProductionDeploymentID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "project does not have a production deployment")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: depl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         time.Hour,
		InstancePermissions: map[string][]runtimeauth.Permission{
			depl.RuntimeInstanceID: {
				// TODO: These are too wide. It needs just ReadObjects and ReadMetrics.
				runtimeauth.ReadInstance,
				runtimeauth.ReadObjects,
				runtimeauth.ReadOLAP,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadProfiling,
				runtimeauth.ReadRepo,
			},
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	return &adminv1.GetProjectResponse{
		Project:              projToDTO(proj),
		ProductionDeployment: deploymentToDTO(depl),
		Jwt:                  jwt,
	}, nil
}

func (s *Server) CreateProject(ctx context.Context, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Find parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get Github installation ID for the repo
	installationID, ok, err := s.admin.GetUserGithubInstallation(ctx, claims.OwnerID(), req.GithubUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get Github installation: %w", err)
	}
	// Check that the user has access to the installation
	if !ok {
		return nil, fmt.Errorf("you have not granted Rill access to %q", req.GithubUrl)
	}

	// TODO: Validate that req.ProductionBranch is an actual branch.

	// TODO: Validate that req.ProductionSlots is an allowed tier for the caller.

	// TODO: Validate that req.ProductionOlapDriver and req.ProductionOlapDsn are acceptable.

	// Create the project
	proj, err := s.admin.CreateProject(ctx, &database.InsertProjectOptions{
		OrganizationID:       org.ID,
		Name:                 req.Name,
		Description:          req.Description,
		Public:               req.Public,
		Region:               req.Region,
		ProductionOLAPDriver: req.ProductionOlapDriver,
		ProductionOLAPDSN:    req.ProductionOlapDsn,
		ProductionSlots:      int(req.ProductionSlots),
		ProductionBranch:     req.ProductionBranch,
		GithubURL:            &req.GithubUrl,
		GithubInstallationID: &installationID,
		ProductionVariables:  req.Variables,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	projectURL, err := url.JoinPath(s.opts.FrontendURL, fmt.Sprintf("-/%s/%s", org.Name, proj.Name))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("project url generation failed with error %s", err.Error()))
	}

	return &adminv1.CreateProjectResponse{
		Project:    projToDTO(proj),
		ProjectUrl: projectURL,
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *adminv1.DeleteProjectRequest) (*adminv1.DeleteProjectResponse, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.admin.TeardownProject(ctx, proj)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteProjectResponse{}, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *adminv1.UpdateProjectRequest) (*adminv1.UpdateProjectResponse, error) {
	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// If changing the Github URL, check the caller has access
	if safeStr(proj.GithubURL) != req.GithubUrl {
		_, ok, err := s.admin.GetUserGithubInstallation(ctx, claims.OwnerID(), req.GithubUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to get Github installation: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("you have not granted Rill access to %q", req.GithubUrl)
		}
	}

	var githubURL *string
	if req.GithubUrl != "" {
		githubURL = &req.GithubUrl
	}

	proj, err = s.admin.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Description:            req.Description,
		Public:                 req.Public,
		ProductionBranch:       req.ProductionBranch,
		ProductionVariables:    req.Variables,
		GithubURL:              githubURL,
		GithubInstallationID:   proj.GithubInstallationID,
		ProductionDeploymentID: proj.ProductionDeploymentID,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.UpdateProjectResponse{
		Project: projToDTO(proj),
	}, nil
}

func projToDTO(p *database.Project) *adminv1.Project {
	return &adminv1.Project{
		Id:                     p.ID,
		Name:                   p.Name,
		Description:            p.Description,
		Public:                 p.Public,
		Region:                 p.Region,
		ProductionOlapDriver:   p.ProductionOLAPDriver,
		ProductionOlapDsn:      p.ProductionOLAPDSN,
		ProductionSlots:        int64(p.ProductionSlots),
		ProductionBranch:       p.ProductionBranch,
		GithubUrl:              safeStr(p.GithubURL),
		ProductionDeploymentId: safeStr(p.ProductionDeploymentID),
		CreatedOn:              timestamppb.New(p.CreatedOn),
		UpdatedOn:              timestamppb.New(p.UpdatedOn),
		Variables:              p.ProductionVariables,
	}
}

func deploymentToDTO(d *database.Deployment) *adminv1.Deployment {
	var s adminv1.DeploymentStatus
	switch d.Status {
	case database.DeploymentStatusUnspecified:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_UNSPECIFIED
	case database.DeploymentStatusPending:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_PENDING
	case database.DeploymentStatusOK:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_OK
	case database.DeploymentStatusReconciling:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RECONCILING
	case database.DeploymentStatusError:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERROR
	default:
		panic(fmt.Errorf("unhandled deployment status %d", d.Status))
	}

	return &adminv1.Deployment{
		Id:                d.ID,
		ProjectId:         d.ProjectID,
		Slots:             int64(d.Slots),
		Branch:            d.Branch,
		RuntimeHost:       d.RuntimeHost,
		RuntimeInstanceId: d.RuntimeInstanceID,
		Status:            s,
		Logs:              d.Logs,
		CreatedOn:         timestamppb.New(d.CreatedOn),
		UpdatedOn:         timestamppb.New(d.UpdatedOn),
	}
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
