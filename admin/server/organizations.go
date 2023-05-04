package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListOrganizations(ctx context.Context, req *adminv1.ListOrganizationsRequest) (*adminv1.ListOrganizationsResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	pageSize := validPageSize(req.PageSize)

	orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, claims.OwnerID(), token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(orgs) >= pageSize {
		nextToken = marshalPageToken(orgs[len(orgs)-1].Name)
	}

	pbs := make([]*adminv1.Organization, len(orgs))
	for i, org := range orgs {
		pbs[i] = organizationToDTO(org)
	}

	return &adminv1.ListOrganizationsResponse{Organizations: pbs, NextPageToken: nextToken}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *adminv1.GetOrganizationRequest) (*adminv1.GetOrganizationResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrg {
		// check if the org has any public projects, this works for anonymous users as well
		hasPublicProject, err := s.admin.DB.CheckOrganizationHasPublicProjects(ctx, org.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// these are the permissions for public and for outside members
		publicPermissions := &adminv1.OrganizationPermissions{ReadOrg: true, ReadProjects: true}
		if hasPublicProject {
			return &adminv1.GetOrganizationResponse{
				Organization: organizationToDTO(org),
				Permissions:  publicPermissions,
			}, nil
		}
		// check if the user is outside members of a project in the org
		if claims.OwnerType() == auth.OwnerTypeUser {
			exists, err := s.admin.DB.CheckOrganizationHasOutsideUser(ctx, org.ID, claims.OwnerID())
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if exists {
				return &adminv1.GetOrganizationResponse{
					Organization: organizationToDTO(org),
					Permissions:  publicPermissions,
				}, nil
			}
		}
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org")
	}

	return &adminv1.GetOrganizationResponse{
		Organization: organizationToDTO(org),
		Permissions:  claims.OrganizationPermissions(ctx, org.ID),
	}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// check single user org limit for this user
	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	count, err := s.admin.DB.CountSingleuserOrganizationsForMemberUser(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user.QuotaSingleuserOrgs >= 0 && count >= user.QuotaSingleuserOrgs {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: you can only create %d single-user orgs", user.QuotaSingleuserOrgs)
	}

	org, err := s.admin.CreateOrganizationForUser(ctx, user.ID, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: organizationToDTO(org),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org")
	}

	err = s.admin.DB.DeleteOrganization(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteOrganizationResponse{}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *adminv1.UpdateOrganizationRequest) (*adminv1.UpdateOrganizationResponse, error) {
	org, err := s.admin.DB.FindOrganization(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, req.Id, &database.UpdateOrganizationOptions{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UpdateOrganizationResponse{
		Organization: organizationToDTO(org),
	}, nil
}

func (s *Server) ListOrganizationMembers(ctx context.Context, req *adminv1.ListOrganizationMembersRequest) (*adminv1.ListOrganizationMembersResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	members, err := s.admin.DB.FindOrganizationMemberUsers(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.Member, len(members))
	for i, user := range members {
		dtos[i] = memberToPB(user)
	}

	return &adminv1.ListOrganizationMembersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListOrganizationInvites(ctx context.Context, req *adminv1.ListOrganizationInvitesRequest) (*adminv1.ListOrganizationInvitesResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	// get pending user invites for this org
	userInvites, err := s.admin.DB.FindOrganizationInvites(ctx, org.ID, token.Val, pageSize)
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

	return &adminv1.ListOrganizationInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) AddOrganizationMember(ctx context.Context, req *adminv1.AddOrganizationMemberRequest) (*adminv1.AddOrganizationMemberResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org members")
	}

	count, err := s.admin.DB.CountInvitesForOrganization(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if org.QuotaOutstandingInvites >= 0 && count >= org.QuotaOutstandingInvites {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can at most have %d outstanding invitations", org.QuotaOutstandingInvites)
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Invite user to join the organization
		invitedBy := ""
		if claims.OwnerType() == auth.OwnerTypeUser {
			invitedBy = claims.OwnerID()
		}
		err = s.admin.InviteUserToOrganization(ctx, &database.InviteUserToOrganizationOptions{
			Email:     req.Email,
			InviterID: invitedBy,
			OrgID:     org.ID,
			RoleID:    role.ID,
			OrgName:   org.Name,
			RoleName:  role.Name,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &adminv1.AddOrganizationMemberResponse{
			PendingSignup: true,
		}, nil
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.admin.DB.InsertUsergroupMember(ctx, *org.AllUsergroupID, user.ID)
	if err != nil {
		if !errors.Is(err, database.ErrNotUnique) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// If the user is already in the all user group, we can ignore the error
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.Email.SendOrganizationAdditionNotification(req.Email, "", org.Name, role.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddOrganizationMemberResponse{
		PendingSignup: false,
	}, nil
}

func (s *Server) RemoveOrganizationMember(ctx context.Context, req *adminv1.RemoveOrganizationMemberRequest) (*adminv1.RemoveOrganizationMemberResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// check if there is a pending invite
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &adminv1.RemoveOrganizationMemberResponse{}, nil
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(errors.Wrap(err, "failed to find organization admin role"))
	}

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(users) == 1 && users[0].ID == user.ID {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user group
	err = s.admin.DB.DeleteUsergroupMember(ctx, *org.AllUsergroupID, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// delete from projects if KeepProjectRoles flag is set
	if !req.KeepProjectRoles {
		err = s.admin.DB.DeleteAllProjectMemberUserForOrganization(ctx, org.ID, user.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveOrganizationMemberResponse{}, nil
}

func (s *Server) SetOrganizationMemberRole(ctx context.Context, req *adminv1.SetOrganizationMemberRoleRequest) (*adminv1.SetOrganizationMemberRoleResponse, error) {
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "user not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		err = s.admin.DB.UpdateOrganizationInviteRole(ctx, invite.ID, role.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &adminv1.SetOrganizationMemberRoleResponse{}, nil
	}

	// Check if the user is the last owner
	if role.Name != database.OrganizationRoleNameAdmin {
		adminRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
		if err != nil {
			panic(errors.Wrap(err, "failed to find organization admin role"))
		}
		// TODO optimize this, may be extract roles during auth token validation
		//  and store as part of the claims and fetch admins only if the user is an admin
		users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, adminRole.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if len(users) == 1 && users[0].ID == user.ID {
			return nil, status.Error(codes.InvalidArgument, "cannot change role of the last owner")
		}
	}

	err = s.admin.DB.UpdateOrganizationMemberUserRole(ctx, org.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.SetOrganizationMemberRoleResponse{}, nil
}

func (s *Server) LeaveOrganization(ctx context.Context, req *adminv1.LeaveOrganizationRequest) (*adminv1.LeaveOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(errors.Wrap(err, "failed to find organization admin role"))
	}

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(users) == 1 && users[0].ID == claims.OwnerID() {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user group
	err = s.admin.DB.DeleteUsergroupMember(ctx, *org.AllUsergroupID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.LeaveOrganizationResponse{}, nil
}

func (s *Server) CreateAutoinviteDomain(ctx context.Context, req *adminv1.CreateAutoinviteDomainRequest) (*adminv1.CreateAutoinviteDomainResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can add autoinvite domain")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.admin.DB.InsertOrganizationAutoinviteDomain(ctx, &database.InsertOrganizationAutoinviteDomainOptions{
		OrgID:     org.ID,
		OrgRoleID: role.ID,
		Domain:    req.Domain,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CreateAutoinviteDomainResponse{}, nil
}

func (s *Server) RemoveAutoinviteDomain(ctx context.Context, req *adminv1.RemoveAutoinviteDomainRequest) (*adminv1.RemoveAutoinviteDomainResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can remove autoinvite domain")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	invite, err := s.admin.DB.FindOrganizationAutoinviteDomain(ctx, org.ID, req.Domain)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "auto invite not found for org %q and domain %q", org.Name, req.Domain)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.DeleteOrganizationAutoinviteDomain(ctx, invite.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveAutoinviteDomainResponse{}, nil
}

func organizationToDTO(o *database.Organization) *adminv1.Organization {
	return &adminv1.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		CreatedOn:   timestamppb.New(o.CreatedOn),
		UpdatedOn:   timestamppb.New(o.UpdatedOn),
	}
}
