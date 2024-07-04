package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUsergroup(ctx context.Context, req *adminv1.CreateUsergroupRequest) (*adminv1.CreateUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.name", req.Name),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org usergroup")
	}

	_, err = s.admin.DB.InsertUsergroup(ctx, &database.InsertUsergroupOptions{
		Name:  req.Name,
		OrgID: org.ID,
	})
	if errors.Is(err, database.ErrNotUnique) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CreateUsergroupResponse{}, nil
}

func (s *Server) GetUsergroup(ctx context.Context, req *adminv1.GetUsergroupRequest) (*adminv1.GetUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	orgRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, usergroup.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if orgRole != nil && !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to get org usergroup")
	}

	projRoles, err := s.admin.DB.FindProjectMemberUsergroupRoles(ctx, usergroup.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, role := range projRoles {
		if !claims.ProjectPermissions(ctx, org.ID, role.ProjectID).ReadProjectMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to get org usergroup")
		}
	}

	return &adminv1.GetUsergroupResponse{
		Usergroup: usergroupToPB(usergroup, orgRole, projRoles),
	}, nil
}

func (s *Server) RenameUsergroup(ctx context.Context, req *adminv1.RenameUsergroupRequest) (*adminv1.RenameUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.name", req.Name),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot rename all-users group")
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to rename org usergroup")
	}

	_, err = s.admin.DB.UpdateUsergroupName(ctx, req.Name, usergroup.ID)
	if errors.Is(err, database.ErrNotUnique) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RenameUsergroupResponse{}, nil
}

func (s *Server) ListUsergroups(ctx context.Context, req *adminv1.ListUsergroupsRequest) (*adminv1.ListUsergroupsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list org usergroups")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	usergroups, err := s.admin.DB.FindOrganizationMemberUsergroups(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orgRoles, err := s.admin.DB.FindOrganizationAllMemberUsergroupRoles(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orgRolesMap := make(map[string]*database.UsergroupOrgRole)
	for _, orgRole := range orgRoles {
		orgRolesMap[orgRole.UsergroupID] = orgRole
	}

	projRoles, err := s.admin.DB.FindUsergroupAllProjectOrganizationRoles(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	projRolesMap := make(map[string][]*database.UsergroupProjectRole)
	for _, projRole := range projRoles {
		projRolesMap[projRole.UsergroupID] = append(projRolesMap[projRole.UsergroupID], projRole)
	}

	nextToken := ""
	if len(usergroups) >= pageSize {
		nextToken = marshalPageToken(usergroups[len(usergroups)-1].Name)
	}

	dtos := make([]*adminv1.Usergroup, len(usergroups))
	for i, usergroup := range usergroups {
		orgRole := orgRolesMap[usergroup.ID]
		projectRoles := projRolesMap[usergroup.ID]
		dtos[i] = usergroupToPB(usergroup, orgRole, projectRoles)
	}

	return &adminv1.ListUsergroupsResponse{
		Usergroups:    dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) DeleteUsergroup(ctx context.Context, req *adminv1.DeleteUsergroupRequest) (*adminv1.DeleteUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org usergroup")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot remove all-users group")
	}

	err = s.admin.DB.DeleteUsergroup(ctx, usergroup.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &adminv1.DeleteUsergroupResponse{}, nil
}

func (s *Server) SetOrganizationMemberUsergroupRole(ctx context.Context, req *adminv1.SetOrganizationMemberUsergroupRoleRequest) (*adminv1.SetOrganizationMemberUsergroupRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org usergroup role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot set role for all-users group")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.admin.DB.UpsertOrganizationMemberUsergroup(ctx, usergroup.ID, org.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.SetOrganizationMemberUsergroupRoleResponse{}, nil
}

func (s *Server) RemoveOrganizationMemberUsergroup(ctx context.Context, req *adminv1.RemoveOrganizationMemberUsergroupRequest) (*adminv1.RemoveOrganizationMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke org usergroup role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot revoke role from all-users group")
	}

	err = s.admin.DB.DeleteOrganizationMemberUsergroup(ctx, usergroup.ID, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveOrganizationMemberUsergroupResponse{}, nil
}

func (s *Server) SetProjectMemberUsergroupRole(ctx context.Context, req *adminv1.SetProjectMemberUsergroupRoleRequest) (*adminv1.SetProjectMemberUsergroupRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project usergroup role")
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot set role for all-users group")
	}

	err = s.admin.DB.UpsertProjectMemberUsergroup(ctx, usergroup.ID, proj.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.SetProjectMemberUsergroupRoleResponse{}, nil
}

func (s *Server) RemoveProjectMemberUsergroup(ctx context.Context, req *adminv1.RemoveProjectMemberUsergroupRequest) (*adminv1.RemoveProjectMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.usergroup", req.Usergroup),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke project usergroup role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if org.AllUsergroupID != nil && usergroup.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot revoke role from all-users group")
	}

	err = s.admin.DB.DeleteProjectMemberUsergroup(ctx, usergroup.ID, proj.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveProjectMemberUsergroupResponse{}, nil
}

func (s *Server) AddUsergroupMemberUser(ctx context.Context, req *adminv1.AddUsergroupMemberUserRequest) (*adminv1.AddUsergroupMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && group.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot add member to all-users group")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	orgRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if orgRole != nil {
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to add usergroup members")
		}
	}

	projectRoles, err := s.admin.DB.FindProjectMemberUsergroupRoles(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, role := range projectRoles {
		if !claims.ProjectPermissions(ctx, org.ID, role.ProjectID).ManageProjectMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to add usergroup members")
		}
	}

	err = s.admin.DB.InsertUsergroupMemberUser(ctx, group.ID, user.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotUnique) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, err
	}

	return &adminv1.AddUsergroupMemberUserResponse{}, nil
}

func (s *Server) ListUsergroupMemberUsers(ctx context.Context, req *adminv1.ListUsergroupMemberUsersRequest) (*adminv1.ListUsergroupMemberUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	orgRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if orgRole != nil {
		if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to list usergroup members")
		}
	}

	projectRoles, err := s.admin.DB.FindProjectMemberUsergroupRoles(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, role := range projectRoles {
		if !claims.ProjectPermissions(ctx, org.ID, role.ProjectID).ReadProjectMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to list usergroup members")
		}
	}

	members, err := s.admin.DB.FindUsergroupMembersUsers(ctx, group.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	usergroupMembers := make([]*adminv1.MemberUser, len(members))
	for i, member := range members {
		usergroupMembers[i] = &adminv1.MemberUser{
			UserId:    member.ID,
			UserEmail: member.Email,
			UserName:  member.DisplayName,
		}
	}

	return &adminv1.ListUsergroupMemberUsersResponse{Members: usergroupMembers}, nil
}

func (s *Server) RemoveUsergroupMemberUser(ctx context.Context, req *adminv1.RemoveUsergroupMemberUserRequest) (*adminv1.RemoveUsergroupMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.usergroup", req.Usergroup),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Usergroup, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if org.AllUsergroupID != nil && group.ID == *org.AllUsergroupID {
		return nil, status.Error(codes.InvalidArgument, "cannot remove member from all-users group")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	orgRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if orgRole != nil {
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to remove usergroup members")
		}
	}

	projectRoles, err := s.admin.DB.FindProjectMemberUsergroupRoles(ctx, group.ID, org.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, role := range projectRoles {
		if !claims.ProjectPermissions(ctx, org.ID, role.ProjectID).ManageProjectMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to remove usergroup members")
		}
	}

	err = s.admin.DB.DeleteUsergroupMemberUser(ctx, group.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveUsergroupMemberUserResponse{}, nil
}

func (s *Server) RemoveUsergroupsMemberUser(ctx context.Context, req *adminv1.RemoveUsergroupsMemberUserRequest) (*adminv1.RemoveUsergroupsMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)

	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove usergroup members")
	}

	err = s.admin.DB.DeleteUsergroupsMemberUser(ctx, user.ID, org.ID, *org.AllUsergroupID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveUsergroupsMemberUserResponse{}, nil
}

func usergroupToPB(group *database.Usergroup, orgRole *database.UsergroupOrgRole, projectRoles []*database.UsergroupProjectRole) *adminv1.Usergroup {
	return &adminv1.Usergroup{
		GroupId:      group.ID,
		GroupName:    group.Name,
		CreatedOn:    timestamppb.New(group.CreatedOn),
		UpdatedOn:    timestamppb.New(group.UpdatedOn),
		OrgRole:      orgRoleToPB(orgRole),
		ProjectRoles: projectRolesToPB(projectRoles),
	}
}

func orgRoleToPB(role *database.UsergroupOrgRole) *adminv1.UsergroupOrgRole {
	if role == nil {
		return nil
	}
	return &adminv1.UsergroupOrgRole{
		OrgName: role.OrgName,
		Role:    role.RoleName,
	}
}

func projectRolesToPB(roles []*database.UsergroupProjectRole) []*adminv1.UsergroupProjectRole {
	pbRoles := make([]*adminv1.UsergroupProjectRole, len(roles))
	for i, role := range roles {
		pbRoles[i] = &adminv1.UsergroupProjectRole{
			ProjectName: role.ProjectName,
			Role:        role.RoleName,
		}
	}
	return pbRoles
}
