package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const prodDeplTTL = 14 * 24 * time.Hour

func (s *Server) ListProjectsForOrganization(ctx context.Context, req *connect.Request[adminv1.ListProjectsForOrganizationRequest]) (*connect.Response[adminv1.ListProjectsForOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

	// If user has ManageProjects, return all projects
	claims := auth.GetClaims(ctx)
	var projs []*database.Project
	if claims.OrganizationPermissions(ctx, org.ID).ManageProjects {
		projs, err = s.admin.DB.FindProjectsForOrganization(ctx, org.ID, token.Val, pageSize)
	} else if claims.OwnerType() == auth.OwnerTypeUser {
		// Get projects the user is a (direct or group) member of (note: the user can be a member of a project in the org, without being a member of org - we call this an "outside member")
		// plus all public projects
		projs, err = s.admin.DB.FindProjectsForOrgAndUser(ctx, org.ID, claims.OwnerID(), token.Val, pageSize)
	} else {
		projs, err = s.admin.DB.FindPublicProjectsInOrganization(ctx, org.ID, token.Val, pageSize)
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// If no projects are public, and user is not an outside member of any projects, the projsMap is empty.
	// If additionally, the user is not an org member, return permission denied (instead of an empty slice).
	if len(projs) == 0 && !claims.OrganizationPermissions(ctx, org.ID).ReadProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read projects")
	}

	nextToken := ""
	if len(projs) >= pageSize {
		nextToken = marshalPageToken(projs[len(projs)-1].Name)
	}

	dtos := make([]*adminv1.Project, len(projs))
	for i, p := range projs {
		dtos[i] = s.projToDTO(p, org.Name)
	}

	return connect.NewResponse(&adminv1.ListProjectsForOrganizationResponse{
		Projects:      dtos,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) GetProject(ctx context.Context, req *connect.Request[adminv1.GetProjectRequest]) (*connect.Response[adminv1.GetProjectResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
		attribute.String("args.project", req.Msg.Name),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.OrganizationName, req.Msg.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProject && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	if proj.ProdDeploymentID == nil || !permissions.ReadProd {
		return connect.NewResponse(&adminv1.GetProjectResponse{
			Project:            s.projToDTO(proj, org.Name),
			ProjectPermissions: permissions,
		}), nil
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !permissions.ReadProdStatus {
		depl.Logs = ""
	}

	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: depl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         time.Hour,
		InstancePermissions: map[string][]runtimeauth.Permission{
			depl.RuntimeInstanceID: {
				// TODO: Remove ReadProfiling and ReadRepo (may require frontend changes)
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadProfiling,
				runtimeauth.ReadRepo,
			},
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(depl.ID)

	return connect.NewResponse(&adminv1.GetProjectResponse{
		Project:            s.projToDTO(proj, org.Name),
		ProdDeployment:     deploymentToDTO(depl),
		Jwt:                jwt,
		ProjectPermissions: permissions,
	}), nil
}

func (s *Server) SearchProjectNames(ctx context.Context, req *connect.Request[adminv1.SearchProjectNamesRequest]) (*connect.Response[adminv1.SearchProjectNamesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.pattern", req.Msg.NamePattern),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can search projects")
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

	projectNames, err := s.admin.DB.FindProjectPathsByPattern(ctx, req.Msg.NamePattern, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nextToken := ""
	if len(projectNames) >= pageSize {
		nextToken = marshalPageToken(projectNames[len(projectNames)-1])
	}

	return connect.NewResponse(&adminv1.SearchProjectNamesResponse{
		Names:         projectNames,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) CreateProject(ctx context.Context, req *connect.Request[adminv1.CreateProjectRequest]) (*connect.Response[adminv1.CreateProjectResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
		attribute.String("args.project", req.Msg.Name),
		attribute.String("args.description", req.Msg.Description),
		attribute.Bool("args.public", req.Msg.Public),
		attribute.String("args.region", req.Msg.Region),
		attribute.String("args.prod_olap_driver", req.Msg.ProdOlapDriver),
		attribute.Int64("args.prod_slots", req.Msg.ProdSlots),
		attribute.String("args.sub_path", req.Msg.Subpath),
		attribute.String("args.prod_branch", req.Msg.ProdBranch),
		attribute.String("args.github_url", req.Msg.GithubUrl),
	)

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Find parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create projects")
	}

	// Check projects quota
	count, err := s.admin.DB.CountProjectsForOrganization(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if org.QuotaProjects >= 0 && count >= org.QuotaProjects {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d projects", org.Name, org.QuotaProjects)
	}

	// Check slots per deployment quota
	if org.QuotaSlotsPerDeployment >= 0 && int(req.Msg.ProdSlots) > org.QuotaSlotsPerDeployment {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can't provision more than %d slots per deployment", org.QuotaSlotsPerDeployment)
	}

	// Check per project deployments and slots limit
	stats, err := s.admin.DB.CountDeploymentsForOrganization(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if org.QuotaDeployments >= 0 && stats.Deployments >= org.QuotaDeployments {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d deployments", org.Name, org.QuotaDeployments)
	}
	if org.QuotaSlotsTotal >= 0 && stats.Slots+int(req.Msg.ProdSlots) > org.QuotaSlotsTotal {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d total slots", org.Name, org.QuotaSlotsTotal)
	}

	// Check Github app is installed and caller has access on the repo
	installationID, err := s.getAndCheckGithubInstallationID(ctx, req.Msg.GithubUrl, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	// Add prod TTL as 7 days if not a public project else infinite
	var prodTTL *int64
	if !req.Msg.Public {
		tmp := int64(prodDeplTTL.Seconds())
		prodTTL = &tmp
	}

	// Create the project
	proj, err := s.admin.CreateProject(ctx, org, claims.OwnerID(), &database.InsertProjectOptions{
		OrganizationID:       org.ID,
		Name:                 req.Msg.Name,
		Description:          req.Msg.Description,
		Public:               req.Msg.Public,
		Region:               req.Msg.Region,
		ProdOLAPDriver:       req.Msg.ProdOlapDriver,
		ProdOLAPDSN:          req.Msg.ProdOlapDsn,
		ProdSlots:            int(req.Msg.ProdSlots),
		Subpath:              req.Msg.Subpath,
		ProdBranch:           req.Msg.ProdBranch,
		GithubURL:            &req.Msg.GithubUrl,
		GithubInstallationID: &installationID,
		ProdVariables:        req.Msg.Variables,
		ProdTTLSeconds:       prodTTL,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&adminv1.CreateProjectResponse{
		Project: s.projToDTO(proj, org.Name),
	}), nil
}

func (s *Server) DeleteProject(ctx context.Context, req *connect.Request[adminv1.DeleteProjectRequest]) (*connect.Response[adminv1.DeleteProjectResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
		attribute.String("args.project", req.Msg.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.OrganizationName, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	err = s.admin.TeardownProject(ctx, proj)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&adminv1.DeleteProjectResponse{}), nil
}

func (s *Server) UpdateProject(ctx context.Context, req *connect.Request[adminv1.UpdateProjectRequest]) (*connect.Response[adminv1.UpdateProjectResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
		attribute.String("args.project", req.Msg.Name),
	)
	if req.Msg.Description != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.description", *req.Msg.Description))
	}
	if req.Msg.Region != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.region", *req.Msg.Region))
	}
	if req.Msg.ProdBranch != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.prod_branch", *req.Msg.ProdBranch))
	}
	if req.Msg.GithubUrl != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.github_url", *req.Msg.GithubUrl))
	}
	if req.Msg.Public != nil {
		observability.AddRequestAttributes(ctx, attribute.Bool("args.public", *req.Msg.Public))
	}
	if req.Msg.ProdSlots != nil {
		observability.AddRequestAttributes(ctx, attribute.Int64("args.prod_slots", *req.Msg.ProdSlots))
	}
	if req.Msg.ProdTtlSeconds != nil {
		observability.AddRequestAttributes(ctx, attribute.Int64("args.prod_ttl_seconds", *req.Msg.ProdTtlSeconds))
	}
	if req.Msg.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.Msg.NewName))
	}

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.OrganizationName, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	githubURL := proj.GithubURL
	if req.Msg.GithubUrl != nil {
		// If changing the Github URL, check github app is installed and caller has access on the repo
		if safeStr(proj.GithubURL) != *req.Msg.GithubUrl {
			_, err = s.getAndCheckGithubInstallationID(ctx, *req.Msg.GithubUrl, claims.OwnerID())
			if err != nil {
				return nil, err
			}
			githubURL = req.Msg.GithubUrl
		}
	}

	prodTTLSeconds := valOrDefault(req.Msg.ProdTtlSeconds, *proj.ProdTTLSeconds)
	opts := &database.UpdateProjectOptions{
		Name:                 valOrDefault(req.Msg.NewName, proj.Name),
		Description:          valOrDefault(req.Msg.Description, proj.Description),
		Public:               valOrDefault(req.Msg.Public, proj.Public),
		GithubURL:            githubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdBranch:           valOrDefault(req.Msg.ProdBranch, proj.ProdBranch),
		ProdVariables:        proj.ProdVariables,
		ProdDeploymentID:     proj.ProdDeploymentID,
		ProdSlots:            int(valOrDefault(req.Msg.ProdSlots, int64(proj.ProdSlots))),
		ProdTTLSeconds:       &prodTTLSeconds,
		Region:               valOrDefault(req.Msg.Region, proj.Region),
	}
	proj, err = s.admin.UpdateProject(ctx, proj, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.UpdateProjectResponse{
		Project: s.projToDTO(proj, req.Msg.OrganizationName),
	}), nil
}

func (s *Server) GetProjectVariables(ctx context.Context, req *connect.Request[adminv1.GetProjectVariablesRequest]) (*connect.Response[adminv1.GetProjectVariablesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.OrganizationName),
		attribute.String("args.project", req.Msg.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.OrganizationName, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project variables")
	}

	return connect.NewResponse(&adminv1.GetProjectVariablesResponse{Variables: proj.ProdVariables}), nil
}

func (s *Server) UpdateProjectVariables(ctx context.Context, req *connect.Request[adminv1.UpdateProjectVariablesRequest]) (*connect.Response[adminv1.UpdateProjectVariablesResponse], error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.OrganizationName, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}

	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update project variables")
	}

	proj, err = s.admin.UpdateProjectVariables(ctx, proj, req.Msg.Variables)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "variables updated failed with error %s", err.Error())
	}

	return connect.NewResponse(&adminv1.UpdateProjectVariablesResponse{Variables: proj.ProdVariables}), nil
}

func (s *Server) ListProjectMembers(ctx context.Context, req *connect.Request[adminv1.ListProjectMembersRequest]) (*connect.Response[adminv1.ListProjectMembersResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.project", req.Msg.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.Organization, req.Msg.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read project members")
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

	members, err := s.admin.DB.FindProjectMemberUsers(ctx, proj.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.Member, len(members))
	for i, member := range members {
		dtos[i] = memberToPB(member)
	}

	return connect.NewResponse(&adminv1.ListProjectMembersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) ListProjectInvites(ctx context.Context, req *connect.Request[adminv1.ListProjectInvitesRequest]) (*connect.Response[adminv1.ListProjectInvitesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.project", req.Msg.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.Organization, req.Msg.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read project members")
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

	// get pending user invites for this project
	userInvites, err := s.admin.DB.FindProjectInvites(ctx, proj.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nextToken := ""
	if len(userInvites) >= pageSize {
		nextToken = marshalPageToken(userInvites[len(userInvites)-1].Email)
	}

	invitesDtos := make([]*adminv1.UserInvite, len(userInvites))
	for i, invite := range userInvites {
		invitesDtos[i] = inviteToPB(invite)
	}

	return connect.NewResponse(&adminv1.ListProjectInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) AddProjectMember(ctx context.Context, req *connect.Request[adminv1.AddProjectMemberRequest]) (*connect.Response[adminv1.AddProjectMemberResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.project", req.Msg.Project),
		attribute.String("args.role", req.Msg.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.Organization, req.Msg.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add project members")
	}

	// Check outstanding invites quota
	count, err := s.admin.DB.CountInvitesForOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if org.QuotaOutstandingInvites >= 0 && count >= org.QuotaOutstandingInvites {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q can at most have %d outstanding invitations", org.Name, org.QuotaOutstandingInvites)
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Msg.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var invitedByUserID, invitedByName string
	if claims.OwnerType() == auth.OwnerTypeUser {
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		invitedByUserID = user.ID
		invitedByName = user.DisplayName
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Invite user to join the project
		err := s.admin.DB.InsertProjectInvite(ctx, &database.InsertProjectInviteOptions{
			Email:     req.Msg.Email,
			InviterID: invitedByUserID,
			ProjectID: proj.ID,
			RoleID:    role.ID,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Send invitation email
		err = s.admin.Email.SendProjectInvite(&email.ProjectInvite{
			ToEmail:       req.Msg.Email,
			ToName:        "",
			OrgName:       org.Name,
			ProjectName:   proj.Name,
			RoleName:      role.Name,
			InvitedByName: invitedByName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return connect.NewResponse(&adminv1.AddProjectMemberResponse{
			PendingSignup: true,
		}), nil
	}

	err = s.admin.DB.InsertProjectMemberUser(ctx, proj.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.admin.Email.SendProjectAddition(&email.ProjectAddition{
		ToEmail:       req.Msg.Email,
		ToName:        "",
		OrgName:       org.Name,
		ProjectName:   proj.Name,
		RoleName:      role.Name,
		InvitedByName: invitedByName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.AddProjectMemberResponse{
		PendingSignup: false,
	}), nil
}

func (s *Server) RemoveProjectMember(ctx context.Context, req *connect.Request[adminv1.RemoveProjectMemberRequest]) (*connect.Response[adminv1.RemoveProjectMemberResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.project", req.Msg.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.Organization, req.Msg.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove project members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// check if there is a pending invite
		invite, err := s.admin.DB.FindProjectInvite(ctx, proj.ID, req.Msg.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.DB.DeleteProjectInvite(ctx, invite.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return connect.NewResponse(&adminv1.RemoveProjectMemberResponse{}), nil
	}

	err = s.admin.DB.DeleteProjectMemberUser(ctx, proj.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.RemoveProjectMemberResponse{}), nil
}

func (s *Server) SetProjectMemberRole(ctx context.Context, req *connect.Request[adminv1.SetProjectMemberRoleRequest]) (*connect.Response[adminv1.SetProjectMemberRoleResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.project", req.Msg.Project),
		attribute.String("args.role", req.Msg.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Msg.Organization, req.Msg.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project member roles")
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Msg.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindProjectInvite(ctx, proj.ID, req.Msg.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.DB.UpdateProjectInviteRole(ctx, invite.ID, role.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return connect.NewResponse(&adminv1.SetProjectMemberRoleResponse{}), nil
	}

	err = s.admin.DB.UpdateProjectMemberUserRole(ctx, proj.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.SetProjectMemberRoleResponse{}), nil
}

// getAndCheckGithubInstallationID returns a valid installation ID iff app is installed and user is a collaborator of the repo
func (s *Server) getAndCheckGithubInstallationID(ctx context.Context, githubURL, userID string) (int64, error) {
	// Get Github installation ID for the repo
	installationID, err := s.admin.GetGithubInstallation(ctx, githubURL)
	if err != nil {
		if errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to %q", githubURL)
		}

		return 0, status.Errorf(codes.Internal, "failed to get Github installation: %q", err.Error())
	}

	if installationID == 0 {
		return 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to %q", githubURL)
	}

	// Check that user is a collaborator on the repo
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return 0, status.Error(codes.Internal, err.Error())
	}

	if user.GithubUsername == "" {
		return 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to your Github account")
	}

	_, err = s.admin.LookupGithubRepoForUser(ctx, installationID, githubURL, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			return 0, status.Errorf(codes.PermissionDenied, "you are not collaborator to the repo %q", githubURL)
		}
		return 0, status.Error(codes.Internal, err.Error())
	}

	return installationID, nil
}

func (s *Server) projToDTO(p *database.Project, orgName string) *adminv1.Project {
	frontendURL, _ := url.JoinPath(s.opts.FrontendURL, orgName, p.Name)

	return &adminv1.Project{
		Id:               p.ID,
		Name:             p.Name,
		Description:      p.Description,
		Public:           p.Public,
		OrgId:            p.OrganizationID,
		OrgName:          orgName,
		Region:           p.Region,
		ProdOlapDriver:   p.ProdOLAPDriver,
		ProdOlapDsn:      p.ProdOLAPDSN,
		ProdSlots:        int64(p.ProdSlots),
		ProdBranch:       p.ProdBranch,
		Subpath:          p.Subpath,
		GithubUrl:        safeStr(p.GithubURL),
		ProdDeploymentId: safeStr(p.ProdDeploymentID),
		ProdTtlSeconds:   safeInt64(p.ProdTTLSeconds),
		FrontendUrl:      frontendURL,
		CreatedOn:        timestamppb.New(p.CreatedOn),
		UpdatedOn:        timestamppb.New(p.UpdatedOn),
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

func safeInt64(s *int64) int64 {
	if s == nil {
		return 0
	}
	return *s
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
