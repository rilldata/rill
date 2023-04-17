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

	orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs := make([]*adminv1.Organization, len(orgs))
	for i, org := range orgs {
		pbs[i] = organizationToDTO(org)
	}

	return &adminv1.ListOrganizationsResponse{Organizations: pbs}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *adminv1.GetOrganizationRequest) (*adminv1.GetOrganizationResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ReadOrg) {
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
			exists, err := s.admin.DB.CheckOrganizationProjectsHasMemberUser(ctx, org.ID, claims.OwnerID())
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

	permissions := claims.OrganizationPermissions(ctx, org.ID)

	return &adminv1.GetOrganizationResponse{
		Organization: organizationToDTO(org),
		Permissions:  permissions,
	}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.CreateOrganizationForUser(ctx, claims.OwnerID(), req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: organizationToDTO(org),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrg) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org")
	}

	err = s.admin.DB.DeleteOrganization(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteOrganizationResponse{}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *adminv1.UpdateOrganizationRequest) (*adminv1.UpdateOrganizationResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrg) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, &database.UpdateOrganizationOptions{
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
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ReadOrgMembers) {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	members, err := s.admin.DB.FindOrganizationMemberUsers(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dtos := make([]*adminv1.Member, len(members))
	for i, user := range members {
		dtos[i] = memberToPB(user)
	}

	// get pending user invites for this org
	userInvites, err := s.admin.DB.FindOrganizationMemberInvitations(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	invitesDtos := make([]*adminv1.UserInvite, len(userInvites))
	for _, invite := range userInvites {
		invitesDtos = append(invitesDtos, inviteToPB(invite))
	}

	return &adminv1.ListOrganizationMembersResponse{
		Members: dtos,
		Invites: invitesDtos,
	}, nil
}

func (s *Server) AddOrganizationMember(ctx context.Context, req *adminv1.AddOrganizationMemberRequest) (*adminv1.AddOrganizationMemberResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrgMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
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
		err = s.admin.InviteUserToOrganization(ctx, req.Email, invitedBy, org.ID, role.ID, org.Name, role.Name)
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
		if errors.Is(err, database.ErrNotUnique) {
			return nil, status.Error(codes.InvalidArgument, "user already member of org")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.InsertUserInUsergroup(ctx, user.ID, *org.AllUsergroupID)
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

	return &adminv1.AddOrganizationMemberResponse{
		PendingSignup: false,
	}, nil
}

func (s *Server) RemoveOrganizationMember(ctx context.Context, req *adminv1.RemoveOrganizationMemberRequest) (*adminv1.RemoveOrganizationMemberResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrgMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// check if there is a pending invite
			invite, err := s.admin.DB.FindOrganizationMemberUserInvitation(ctx, org.ID, req.Email)
			if err != nil {
				if errors.Is(err, database.ErrNotFound) {
					return nil, status.Error(codes.InvalidArgument, "user not found")
				}
				return nil, status.Error(codes.Internal, err.Error())
			}
			err = s.admin.DB.DeleteOrganizationMemberUserInvitation(ctx, invite.ID)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return &adminv1.RemoveOrganizationMemberResponse{}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationAdminRoleName)
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
	err = s.admin.DB.DeleteUserFromUsergroup(ctx, user.ID, *org.AllUsergroupID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RemoveOrganizationMemberResponse{}, nil
}

func (s *Server) SetOrganizationMemberRole(ctx context.Context, req *adminv1.SetOrganizationMemberRoleRequest) (*adminv1.SetOrganizationMemberRoleResponse, error) {
	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrgMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check if the user is the last owner
	if role.Name != database.OrganizationAdminRoleName {
		adminRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationAdminRoleName)
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
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ManageOrgMembers) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationAdminRoleName)
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
	err = s.admin.DB.DeleteUserFromUsergroup(ctx, claims.OwnerID(), *org.AllUsergroupID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.LeaveOrganizationResponse{}, nil
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
