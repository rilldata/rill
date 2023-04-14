package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListProjectsForOrganization(ctx context.Context, req *adminv1.ListProjectsForOrganizationRequest) (*adminv1.ListProjectsForOrganizationResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	orgProjects := map[string]*database.Project{}
	// add public projects
	publicProjects, err := s.admin.DB.FindPublicProjectsInOrganization(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, proj := range publicProjects {
		orgProjects[proj.Name] = proj
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ReadProjects) {
		// check if the user is an outside member of a project in the org
		if claims.OwnerType() == auth.OwnerTypeUser {
			projs, err := s.admin.DB.FindProjectsForProjectMemberUser(ctx, org.ID, claims.OwnerID())
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			for _, proj := range projs {
				orgProjects[proj.Name] = proj
			}
		}
		if len(orgProjects) == 0 {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to read projects")
		}
	} else {
		projs, err := s.admin.DB.FindProjectsForOrganization(ctx, org.ID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		for _, proj := range projs {
			orgProjects[proj.Name] = proj
		}
	}

	dtos := make([]*adminv1.Project, len(orgProjects))
	i := 0
	for _, proj := range orgProjects {
		dtos[i] = projToDTO(proj, org.Name)
		i++
	}
	// sort dtos by name
	sort.Slice(dtos, func(i, j int) bool { return dtos[i].Name < dtos[j].Name })

	return &adminv1.ListProjectsForOrganizationResponse{Projects: dtos}, nil
}

func (s *Server) ListProjectsForOrganizationAndGithubURL(ctx context.Context, req *adminv1.ListProjectsForOrganizationAndGithubURLRequest) (*adminv1.ListProjectsForOrganizationResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ReadProjects) {
		return nil, status.Errorf(codes.PermissionDenied, "does not have permission to read projects in org %s", req.OrganizationName)
	}

	projects, err := s.admin.DB.FindProjectsByOrgIDAndGithubURL(ctx, org.ID, req.GithubUrl)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "project with github url %s not found in org %s", req.GithubUrl, req.OrganizationName)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	accessibleProjects := make([]*adminv1.Project, 0)
	for _, p := range projects {
		if claims.CanProject(ctx, p.ID, auth.ReadProject) {
			accessibleProjects = append(accessibleProjects, projToDTO(p, org.Name))
		}
	}

	return &adminv1.ListProjectsForOrganizationResponse{Projects: accessibleProjects}, nil
}

func (s *Server) GetProject(ctx context.Context, req *adminv1.GetProjectRequest) (*adminv1.GetProjectResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ReadProjects, proj.ID, auth.ReadProject) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	if proj.ProductionDeploymentID == nil {
		return &adminv1.GetProjectResponse{
			Project: projToDTO(proj, org.Name),
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
		Project:              projToDTO(proj, org.Name),
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.CreateProjects) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create projects")
	}

	// check github app is installed and caller has access on the repo
	installationID, err := s.fetchInstallationID(ctx, req.GithubUrl, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	// TODO: Validate that req.ProductionBranch is an actual branch.

	// TODO: Validate that req.ProductionSlots is an allowed tier for the caller.

	// TODO: Validate that req.ProductionOlapDriver and req.ProductionOlapDsn are acceptable.

	// Create the project
	proj, err := s.admin.CreateProject(ctx, &database.InsertProjectOptions{
		OrganizationID:       org.ID,
		Name:                 req.Name,
		UserID:               claims.OwnerID(),
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

	projectURL, err := url.JoinPath(s.opts.FrontendURL, org.Name, proj.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("project url generation failed with error %s", err.Error()))
	}

	return &adminv1.CreateProjectResponse{
		Project:    projToDTO(proj, org.Name),
		ProjectUrl: projectURL,
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *adminv1.DeleteProjectRequest) (*adminv1.DeleteProjectResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ManageProjects, proj.ID, auth.ManageProject) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	err = s.admin.TeardownProject(ctx, proj)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteProjectResponse{}, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *adminv1.UpdateProjectRequest) (*adminv1.UpdateProjectResponse, error) {
	claims := auth.GetClaims(ctx)

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ManageProjects, proj.ID, auth.ManageProject) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	// If changing the Github URL, check github app is installed and caller has access on the repo
	if safeStr(proj.GithubURL) != req.GithubUrl {
		_, err = s.fetchInstallationID(ctx, req.GithubUrl, claims.OwnerID())
		if err != nil {
			return nil, err
		}
	}

	githubURL := proj.GithubURL
	if req.GithubUrl != "" {
		githubURL = &req.GithubUrl
	}

	proj, err = s.admin.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Description:            req.Description,
		Public:                 req.Public,
		ProductionBranch:       req.ProductionBranch,
		ProductionVariables:    proj.ProductionVariables,
		GithubURL:              githubURL,
		GithubInstallationID:   proj.GithubInstallationID,
		ProductionDeploymentID: proj.ProductionDeploymentID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UpdateProjectResponse{
		Project: projToDTO(proj, req.OrganizationName),
	}, nil
}

func (s *Server) ListProjectMembers(ctx context.Context, req *adminv1.ListProjectMembersRequest) (*adminv1.ListProjectMembersResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "project not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ReadOrgMembers, proj.ID, auth.ReadProjectMembers) {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read project members")
	}

	members, err := s.admin.DB.FindProjectMemberUsers(ctx, proj.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dtos := make([]*adminv1.Member, len(members))
	for i, member := range members {
		dtos[i] = memberToPB(member)
	}

	return &adminv1.ListProjectMembersResponse{Members: dtos}, nil
}

func (s *Server) AddProjectMember(ctx context.Context, req *adminv1.AddProjectMemberRequest) (*adminv1.AddProjectMemberResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ManageOrgMembers, proj.ID, auth.ManageProjectMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add project members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// Create phantom user
		// TODO: Replace by an invite-based approach
		user, err = s.admin.CreateOrUpdateUser(ctx, req.Email, "", "")
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.InsertProjectMemberUser(ctx, proj.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddProjectMemberResponse{}, nil
}

func (s *Server) RemoveProjectMember(ctx context.Context, req *adminv1.RemoveProjectMemberRequest) (*adminv1.RemoveProjectMemberResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ManageOrgMembers, proj.ID, auth.ManageProjectMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove project members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.DeleteProjectMemberUser(ctx, proj.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveProjectMemberResponse{}, nil
}

func (s *Server) SetProjectMemberRole(ctx context.Context, req *adminv1.SetProjectMemberRoleRequest) (*adminv1.SetProjectMemberRoleResponse, error) {
	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "proj not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Can(ctx, proj.OrganizationID, auth.ManageOrgMembers, proj.ID, auth.ManageProjectMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project member roles")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.UpdateProjectMemberUserRole(ctx, proj.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SetProjectMemberRoleResponse{}, nil
}

func (s *Server) GetProjectVariables(ctx context.Context, req *adminv1.GetProjectVariablesRequest) (*adminv1.GetProjectVariablesResponse, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.CanProject(ctx, proj.ID, auth.ManageProject) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project variables")
	}

	return &adminv1.GetProjectVariablesResponse{Variables: proj.ProductionVariables}, nil
}

func (s *Server) UpdateProjectVariables(ctx context.Context, req *adminv1.UpdateProjectVariablesRequest) (*adminv1.UpdateProjectVariablesResponse, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "proj not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.CanProject(ctx, proj.ID, auth.ManageProject) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update project variables")
	}

	proj, err = s.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Description:            proj.Description,
		Public:                 proj.Public,
		ProductionBranch:       proj.ProductionBranch,
		GithubURL:              proj.GithubURL,
		GithubInstallationID:   proj.GithubInstallationID,
		ProductionDeploymentID: proj.ProductionDeploymentID,
		ProductionVariables:    req.Variables,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "variables updated failed with error %s", err.Error())
	}

	return &adminv1.UpdateProjectVariablesResponse{Variables: proj.ProductionVariables}, nil
}

// fetchInstallationID returns a valid installation ID iff app is installed and user is a collaborator of the repo
func (s *Server) fetchInstallationID(ctx context.Context, githubURL, userID string) (int64, error) {
	// Get Github installation ID for the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, githubURL)
	if err != nil {
		if errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to %q", githubURL)
		}

		return 0, status.Errorf(codes.Internal, "failed to get Github installation: %q", err.Error())
	}

	if installationID == 0 {
		return 0, status.Errorf(codes.Internal, "you have not granted Rill access to %q", githubURL)
	}

	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return 0, status.Error(codes.Internal, err.Error())
	}

	// check that user is a collaborator on the repo
	_, err = s.admin.LookupGithubRepoForUser(ctx, installationID, githubURL, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			return 0, status.Errorf(codes.PermissionDenied, "you are not collaborator to the repo %q", githubURL)
		}

		return 0, status.Error(codes.Internal, err.Error())
	}

	return installationID, nil
}

func projToDTO(p *database.Project, orgName string) *adminv1.Project {
	return &adminv1.Project{
		Id:                     p.ID,
		Name:                   p.Name,
		Description:            p.Description,
		Public:                 p.Public,
		OrgId:                  p.OrganizationID,
		OrgName:                orgName,
		Region:                 p.Region,
		ProductionOlapDriver:   p.ProductionOLAPDriver,
		ProductionOlapDsn:      p.ProductionOLAPDSN,
		ProductionSlots:        int64(p.ProductionSlots),
		ProductionBranch:       p.ProductionBranch,
		GithubUrl:              safeStr(p.GithubURL),
		ProductionDeploymentId: safeStr(p.ProductionDeploymentID),
		CreatedOn:              timestamppb.New(p.CreatedOn),
		UpdatedOn:              timestamppb.New(p.UpdatedOn),
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
