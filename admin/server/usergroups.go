package server

import (
	"context"
	"errors"
	"slices"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListUsergroupsForOrganizationAndUser(ctx context.Context, req *adminv1.ListUsergroupsForOrganizationAndUserRequest) (*adminv1.ListUsergroupsForOrganizationAndUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.user_id", req.UserId),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	if !auth.GetClaims(ctx).OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list user groups for user")
	}

	pageToken, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	usergroups, err := s.admin.DB.FindUsergroupsForOrganizationAndUser(ctx, org.ID, req.UserId, pageToken.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(usergroups) >= pageSize {
		nextToken = marshalPageToken(usergroups[len(usergroups)-1].Name)
	}

	dtos := make([]*adminv1.Usergroup, len(usergroups))
	for i, group := range usergroups {
		dtos[i] = usergroupToPB(group)
	}

	return &adminv1.ListUsergroupsForOrganizationAndUserResponse{
		Usergroups:    dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) CreateUsergroup(ctx context.Context, req *adminv1.CreateUsergroupRequest) (*adminv1.CreateUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.name", req.Name),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org user group")
	}

	grp, err := s.admin.DB.InsertUsergroup(ctx, &database.InsertUsergroupOptions{
		Name:    req.Name,
		OrgID:   org.ID,
		Managed: false,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateUsergroupResponse{
		Usergroup: usergroupToPB(grp),
	}, nil
}

func (s *Server) GetUsergroup(ctx context.Context, req *adminv1.GetUsergroupRequest) (*adminv1.GetUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to get org user group")
	}

	return &adminv1.GetUsergroupResponse{
		Usergroup: usergroupToPB(usergroup),
	}, nil
}

func (s *Server) RenameUsergroup(ctx context.Context, req *adminv1.RenameUsergroupRequest) (*adminv1.RenameUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.name", req.Name),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to rename org user group")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	_, err = s.admin.DB.UpdateUsergroupName(ctx, req.Name, usergroup.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RenameUsergroupResponse{}, nil
}

func (s *Server) EditUsergroup(ctx context.Context, req *adminv1.EditUsergroupRequest) (*adminv1.EditUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.description", req.Description),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to rename org user group")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	_, err = s.admin.DB.UpdateUsergroupDescription(ctx, req.Description, usergroup.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.EditUsergroupResponse{}, nil
}

func (s *Server) ListOrganizationMemberUsergroups(ctx context.Context, req *adminv1.ListOrganizationMemberUsergroupsRequest) (*adminv1.ListOrganizationMemberUsergroupsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list org user groups")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	var roleID string
	if req.Role != "" {
		role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
		if err != nil {
			return nil, err
		}
		roleID = role.ID
	}

	members, err := s.admin.DB.FindOrganizationMemberUsergroups(ctx, org.ID, roleID, req.IncludeCounts, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Name)
	}

	dtos := make([]*adminv1.MemberUsergroup, len(members))
	for i, usergroup := range members {
		dtos[i] = memberUsergroupToPB(usergroup)
	}

	return &adminv1.ListOrganizationMemberUsergroupsResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListProjectMemberUsergroups(ctx context.Context, req *adminv1.ListProjectMemberUsergroupsRequest) (*adminv1.ListProjectMemberUsergroupsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list project user groups")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	var roleID string
	if req.Role != "" {
		role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
		if err != nil {
			return nil, err
		}
		roleID = role.ID
	}

	members, err := s.admin.DB.FindProjectMemberUsergroups(ctx, proj.ID, roleID, req.IncludeCounts, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Name)
	}

	dtos := make([]*adminv1.MemberUsergroup, len(members))
	for i, m := range members {
		dtos[i] = memberUsergroupToPB(m)
	}

	return &adminv1.ListProjectMemberUsergroupsResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) DeleteUsergroup(ctx context.Context, req *adminv1.DeleteUsergroupRequest) (*adminv1.DeleteUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org user group")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	err = s.admin.DB.DeleteUsergroup(ctx, usergroup.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.DeleteUsergroupResponse{}, nil
}

func (s *Server) AddOrganizationMemberUsergroup(ctx context.Context, req *adminv1.AddOrganizationMemberUsergroupRequest) (*adminv1.AddOrganizationMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org user group role")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	err = s.admin.DB.InsertOrganizationMemberUsergroup(ctx, usergroup.ID, usergroup.OrgID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.AddOrganizationMemberUsergroupResponse{}, nil
}

func (s *Server) SetOrganizationMemberUsergroupRole(ctx context.Context, req *adminv1.SetOrganizationMemberUsergroupRoleRequest) (*adminv1.SetOrganizationMemberUsergroupRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org user group role")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	currentRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, usergroup.ID, usergroup.OrgID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if currentRole != nil && currentRole.Admin && !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin role")
	}

	err = s.admin.DB.UpdateOrganizationMemberUsergroup(ctx, usergroup.ID, usergroup.OrgID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SetOrganizationMemberUsergroupRoleResponse{}, nil
}

func (s *Server) RemoveOrganizationMemberUsergroup(ctx context.Context, req *adminv1.RemoveOrganizationMemberUsergroupRequest) (*adminv1.RemoveOrganizationMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke org user group role")
	}

	if usergroup.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	currentRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, usergroup.ID, usergroup.OrgID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if currentRole != nil && currentRole.Admin && !claims.OrganizationPermissions(ctx, usergroup.OrgID).ManageOrgAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin role")
	}

	err = s.admin.DB.DeleteOrganizationMemberUsergroup(ctx, usergroup.ID, usergroup.OrgID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveOrganizationMemberUsergroupResponse{}, nil
}

func (s *Server) AddProjectMemberUsergroup(ctx context.Context, req *adminv1.AddProjectMemberUsergroupRequest) (*adminv1.AddProjectMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add project user group role")
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.InsertProjectMemberUsergroup(ctx, usergroup.ID, proj.ID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.AddProjectMemberUsergroupResponse{}, nil
}

func (s *Server) SetProjectMemberUsergroupRole(ctx context.Context, req *adminv1.SetProjectMemberUsergroupRoleRequest) (*adminv1.SetProjectMemberUsergroupRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.usergroup", req.Usergroup),
		attribute.String("args.role", req.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project user group role")
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	currentRole, err := s.admin.DB.FindProjectMemberUsergroupRole(ctx, usergroup.ID, proj.ID)
	if err != nil {
		return nil, err
	}
	if currentRole.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin role")
	}

	err = s.admin.DB.UpdateProjectMemberUsergroup(ctx, usergroup.ID, proj.ID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SetProjectMemberUsergroupRoleResponse{}, nil
}

func (s *Server) RemoveProjectMemberUsergroup(ctx context.Context, req *adminv1.RemoveProjectMemberUsergroupRequest) (*adminv1.RemoveProjectMemberUsergroupResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.usergroup", req.Usergroup),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke project user group role")
	}

	usergroup, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	currentRole, err := s.admin.DB.FindProjectMemberUsergroupRole(ctx, usergroup.ID, proj.ID)
	if err != nil {
		return nil, err
	}
	if currentRole.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin role")
	}

	err = s.admin.DB.DeleteProjectMemberUsergroup(ctx, usergroup.ID, proj.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveProjectMemberUsergroupResponse{}, nil
}

func (s *Server) AddUsergroupMemberUser(ctx context.Context, req *adminv1.AddUsergroupMemberUserRequest) (*adminv1.AddUsergroupMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, group.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add user group members")
	}

	if group.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	currentRole, err := s.admin.DB.FindOrganizationMemberUsergroupRole(ctx, group.ID, group.OrgID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if currentRole != nil && currentRole.Admin && !claims.OrganizationPermissions(ctx, group.OrgID).ManageOrgAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to edit a group that has an admin role")
	}
	// NOTE: In theory, the group could be admin on a project that the current user is not admin on.
	// We don't check for that because it's complicated and not a big leak of permissions.

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
		// did not find user, check if there is any pending invite
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, group.OrgID, req.Email)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return nil, err
			}
			// there is no pending invite return error
			return nil, status.Error(codes.FailedPrecondition, "user is not a member of the organization")
		}
		// add group to the invite, dedupe the group ids
		if !slices.Contains(invite.UsergroupIDs, group.ID) {
			invite.UsergroupIDs = append(invite.UsergroupIDs, group.ID)
			err = s.admin.DB.UpdateOrganizationInviteUsergroups(ctx, invite.ID, invite.UsergroupIDs)
			if err != nil {
				return nil, err
			}
		}

		return &adminv1.AddUsergroupMemberUserResponse{}, nil
	}

	isOrgMember, err := s.admin.DB.CheckUserIsAnOrganizationMember(ctx, user.ID, group.OrgID)
	if err != nil {
		return nil, err
	}
	if !isOrgMember {
		return nil, status.Error(codes.FailedPrecondition, "user is not a member of the organization")
	}

	err = s.admin.DB.InsertUsergroupMemberUser(ctx, group.ID, user.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.AddUsergroupMemberUserResponse{}, nil
}

func (s *Server) ListUsergroupMemberUsers(ctx context.Context, req *adminv1.ListUsergroupMemberUsersRequest) (*adminv1.ListUsergroupMemberUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, group.OrgID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list user group members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	members, err := s.admin.DB.FindUsergroupMemberUsers(ctx, group.ID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.UsergroupMemberUser, len(members))
	for i, member := range members {
		dtos[i] = usergroupMemberUserToPB(member)
	}

	return &adminv1.ListUsergroupMemberUsersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) RemoveUsergroupMemberUser(ctx context.Context, req *adminv1.RemoveUsergroupMemberUserRequest) (*adminv1.RemoveUsergroupMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.usergroup", req.Usergroup),
	)

	group, err := s.admin.DB.FindUsergroupByName(ctx, req.Org, req.Usergroup)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, group.OrgID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove user group members")
	}

	if group.Managed {
		return nil, status.Error(codes.InvalidArgument, "cannot edit managed user group")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteUsergroupMemberUser(ctx, group.ID, user.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveUsergroupMemberUserResponse{}, nil
}

func usergroupToPB(group *database.Usergroup) *adminv1.Usergroup {
	return &adminv1.Usergroup{
		GroupId:          group.ID,
		GroupName:        group.Name,
		GroupDescription: group.Description,
		Managed:          group.Managed,
		CreatedOn:        timestamppb.New(group.CreatedOn),
		UpdatedOn:        timestamppb.New(group.UpdatedOn),
	}
}

func memberUsergroupToPB(member *database.MemberUsergroup) *adminv1.MemberUsergroup {
	return &adminv1.MemberUsergroup{
		GroupId:      member.ID,
		GroupName:    member.Name,
		GroupManaged: member.Managed,
		RoleName:     member.RoleName,
		UsersCount:   uint32(member.UsersCount),
		CreatedOn:    timestamppb.New(member.CreatedOn),
		UpdatedOn:    timestamppb.New(member.UpdatedOn),
	}
}
