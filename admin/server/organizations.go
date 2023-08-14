package server

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/pkg/publicemail"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListOrganizations(ctx context.Context, req *connect.Request[adminv1.ListOrganizationsRequest]) (*connect.Response[adminv1.ListOrganizationsResponse], error) {
	// // Check the request is made by an authenticated user
	// claims := auth.GetClaims(ctx)
	// if claims.OwnerType() != auth.OwnerTypeUser {
	// 	return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	// }

	// token, err := unmarshalPageToken(req.Msg.PageToken)
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }
	pageSize := validPageSize(req.Msg.PageSize)

	// orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, claims.OwnerID(), token.Val, pageSize)
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	orgs, err := s.admin.DB.FindOrganizations(ctx, "all", 4)
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

	return connect.NewResponse(&adminv1.ListOrganizationsResponse{Organizations: pbs, NextPageToken: nextToken}), nil
}

func (s *Server) GetOrganization(ctx context.Context, req *connect.Request[adminv1.GetOrganizationRequest]) (*connect.Response[adminv1.GetOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Msg.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrg && !claims.Superuser(ctx) {
		// check if the org has any public projects, this works for anonymous users as well
		hasPublicProject, err := s.admin.DB.CheckOrganizationHasPublicProjects(ctx, org.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// these are the permissions for public and for outside members
		publicPermissions := &adminv1.OrganizationPermissions{ReadOrg: true, ReadProjects: true}
		if hasPublicProject {
			return connect.NewResponse(&adminv1.GetOrganizationResponse{
				Organization: organizationToDTO(org),
				Permissions:  publicPermissions,
			}), nil
		}
		// check if the user is outside members of a project in the org
		if claims.OwnerType() == auth.OwnerTypeUser {
			exists, err := s.admin.DB.CheckOrganizationHasOutsideUser(ctx, org.ID, claims.OwnerID())
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if exists {
				return connect.NewResponse(&adminv1.GetOrganizationResponse{
					Organization: organizationToDTO(org),
					Permissions:  publicPermissions,
				}), nil
			}
		}
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org")
	}

	return connect.NewResponse(&adminv1.GetOrganizationResponse{
		Organization: organizationToDTO(org),
		Permissions:  claims.OrganizationPermissions(ctx, org.ID),
	}), nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *connect.Request[adminv1.CreateOrganizationRequest]) (*connect.Response[adminv1.CreateOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Name),
		attribute.String("args.description", req.Msg.Description),
	)

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

	org, err := s.admin.CreateOrganizationForUser(ctx, user.ID, req.Msg.Name, req.Msg.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&adminv1.CreateOrganizationResponse{
		Organization: organizationToDTO(org),
	}), nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *connect.Request[adminv1.DeleteOrganizationRequest]) (*connect.Response[adminv1.DeleteOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Msg.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org")
	}

	err = s.admin.DB.DeleteOrganization(ctx, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&adminv1.DeleteOrganizationResponse{}), nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *connect.Request[adminv1.UpdateOrganizationRequest]) (*connect.Response[adminv1.UpdateOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Msg.Name))
	if req.Msg.Description != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.description", *req.Msg.Description))
	}
	if req.Msg.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.Msg.NewName))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                    valOrDefault(req.Msg.NewName, org.Name),
		Description:             valOrDefault(req.Msg.Description, org.Description),
		QuotaProjects:           org.QuotaProjects,
		QuotaDeployments:        org.QuotaDeployments,
		QuotaSlotsTotal:         org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment: org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites: org.QuotaOutstandingInvites,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.UpdateOrganizationResponse{
		Organization: organizationToDTO(org),
	}), nil
}

func (s *Server) ListOrganizationMembers(ctx context.Context, req *connect.Request[adminv1.ListOrganizationMembersRequest]) (*connect.Response[adminv1.ListOrganizationMembersResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

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

	return connect.NewResponse(&adminv1.ListOrganizationMembersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) ListOrganizationInvites(ctx context.Context, req *connect.Request[adminv1.ListOrganizationInvitesRequest]) (*connect.Response[adminv1.ListOrganizationInvitesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.Msg.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.Msg.PageSize)

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

	return connect.NewResponse(&adminv1.ListOrganizationInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}), nil
}

func (s *Server) AddOrganizationMember(ctx context.Context, req *connect.Request[adminv1.AddOrganizationMemberRequest]) (*connect.Response[adminv1.AddOrganizationMemberResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.role", req.Msg.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
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

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Msg.Role)
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

		// Invite user to join org
		err := s.admin.DB.InsertOrganizationInvite(ctx, &database.InsertOrganizationInviteOptions{
			Email:     req.Msg.Email,
			InviterID: invitedByUserID,
			OrgID:     org.ID,
			RoleID:    role.ID,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Send invitation email
		err = s.admin.Email.SendOrganizationInvite(&email.OrganizationInvite{
			ToEmail:       req.Msg.Email,
			ToName:        "",
			OrgName:       org.Name,
			RoleName:      role.Name,
			InvitedByName: invitedByName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return connect.NewResponse(&adminv1.AddOrganizationMemberResponse{
			PendingSignup: true,
		}), nil
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

	err = s.admin.Email.SendOrganizationAddition(&email.OrganizationAddition{
		ToEmail:       req.Msg.Email,
		ToName:        "",
		OrgName:       org.Name,
		RoleName:      role.Name,
		InvitedByName: invitedByName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.AddOrganizationMemberResponse{
		PendingSignup: false,
	}), nil
}

func (s *Server) RemoveOrganizationMember(ctx context.Context, req *connect.Request[adminv1.RemoveOrganizationMemberRequest]) (*connect.Response[adminv1.RemoveOrganizationMemberResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.Bool("args.keep_project_roles", req.Msg.KeepProjectRoles),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// check if there is a pending invite
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Msg.Email)
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
		return connect.NewResponse(&adminv1.RemoveOrganizationMemberResponse{}), nil
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(err)
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
	if !req.Msg.KeepProjectRoles {
		err = s.admin.DB.DeleteAllProjectMemberUserForOrganization(ctx, org.ID, user.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.RemoveOrganizationMemberResponse{}), nil
}

func (s *Server) SetOrganizationMemberRole(ctx context.Context, req *connect.Request[adminv1.SetOrganizationMemberRoleRequest]) (*connect.Response[adminv1.SetOrganizationMemberRoleResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.role", req.Msg.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Msg.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Msg.Email)
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
		return connect.NewResponse(&adminv1.SetOrganizationMemberRoleResponse{}), nil
	}

	// Check if the user is the last owner
	if role.Name != database.OrganizationRoleNameAdmin {
		adminRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
		if err != nil {
			panic(err)
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

	return connect.NewResponse(&adminv1.SetOrganizationMemberRoleResponse{}), nil
}

func (s *Server) LeaveOrganization(ctx context.Context, req *connect.Request[adminv1.LeaveOrganizationRequest]) (*connect.Response[adminv1.LeaveOrganizationResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(err)
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

	return connect.NewResponse(&adminv1.LeaveOrganizationResponse{}), nil
}

func (s *Server) CreateWhitelistedDomain(ctx context.Context, req *connect.Request[adminv1.CreateWhitelistedDomainRequest]) (*connect.Response[adminv1.CreateWhitelistedDomainResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.domain", req.Msg.Domain),
		attribute.String("args.role", req.Msg.Role),
	)

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !claims.Superuser(ctx) {
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
			return nil, status.Error(codes.PermissionDenied, "only org admins can add whitelisted domain")
		}
		// check if the user's domain matches the whitelist domain
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !strings.HasSuffix(user.Email, "@"+req.Msg.Domain) {
			return nil, status.Error(codes.PermissionDenied, "Domain name doesn’t match verified email domain. Please contact Rill support.")
		}

		if publicemail.IsPublic(req.Msg.Domain) {
			return nil, status.Errorf(codes.InvalidArgument, "Public Domain %s cannot be whitelisted", req.Msg.Domain)
		}
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Msg.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// find existing users belonging to the whitelisted domain to the org
	users, err := s.admin.DB.FindUsersByEmailPattern(ctx, "%@"+req.Msg.Domain, "", math.MaxInt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// filter out users who are already members of the org
	newUsers := make([]*database.User, 0)
	for _, user := range users {
		// check if user is already a member of the org
		exists, err := s.admin.DB.CheckUserIsAnOrganizationMember(ctx, user.ID, org.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !exists {
			newUsers = append(newUsers, user)
		}
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = s.admin.DB.InsertOrganizationWhitelistedDomain(ctx, &database.InsertOrganizationWhitelistedDomainOptions{
		OrgID:     org.ID,
		OrgRoleID: role.ID,
		Domain:    req.Msg.Domain,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, user := range newUsers {
		err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// add to all user group
		err = s.admin.DB.InsertUsergroupMember(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&adminv1.CreateWhitelistedDomainResponse{}), nil
}

func (s *Server) RemoveWhitelistedDomain(ctx context.Context, req *connect.Request[adminv1.RemoveWhitelistedDomainRequest]) (*connect.Response[adminv1.RemoveWhitelistedDomainResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
		attribute.String("args.domain", req.Msg.Domain),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can remove whitelisted domain")
	}

	invite, err := s.admin.DB.FindOrganizationWhitelistedDomain(ctx, org.ID, req.Msg.Domain)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "whitelist not found for org %q and domain %q", org.Name, req.Msg.Domain)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.DeleteOrganizationWhitelistedDomain(ctx, invite.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return connect.NewResponse(&adminv1.RemoveWhitelistedDomainResponse{}), nil
}

func (s *Server) ListWhitelistedDomains(ctx context.Context, req *connect.Request[adminv1.ListWhitelistedDomainsRequest]) (*connect.Response[adminv1.ListWhitelistedDomainsResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Msg.Organization),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.Organization)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "org not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can list whitelisted domain")
	}

	whitelistedDomains, err := s.admin.DB.FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	whitelistedDomainDtos := make([]*adminv1.WhitelistedDomain, len(whitelistedDomains))
	for i, whitelistedDomain := range whitelistedDomains {
		whitelistedDomainDtos[i] = whitelistedDomainToPB(whitelistedDomain)
	}

	return connect.NewResponse(&adminv1.ListWhitelistedDomainsResponse{
		Domains: whitelistedDomainDtos,
	}), nil
}

func (s *Server) SudoUpdateOrganizationQuotas(ctx context.Context, req *connect.Request[adminv1.SudoUpdateOrganizationQuotasRequest]) (*connect.Response[adminv1.SudoUpdateOrganizationQuotasResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Msg.OrgName))
	if req.Msg.Projects != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.projects", int(*req.Msg.Projects)))
	}
	if req.Msg.Deployments != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.deployments", int(*req.Msg.Deployments)))
	}
	if req.Msg.SlotsTotal != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_total", int(*req.Msg.SlotsTotal)))
	}
	if req.Msg.SlotsPerDeployment != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_per_deployment", int(*req.Msg.SlotsPerDeployment)))
	}
	if req.Msg.OutstandingInvites != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.outstanding_invites", int(*req.Msg.OutstandingInvites)))
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage quotas")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Msg.OrgName)
	if err != nil {
		return nil, err
	}

	opts := &database.UpdateOrganizationOptions{
		Name:                    req.Msg.OrgName,
		Description:             org.Description,
		QuotaProjects:           int(valOrDefault(req.Msg.Projects, uint32(org.QuotaProjects))),
		QuotaDeployments:        int(valOrDefault(req.Msg.Deployments, uint32(org.QuotaDeployments))),
		QuotaSlotsTotal:         int(valOrDefault(req.Msg.SlotsTotal, uint32(org.QuotaSlotsTotal))),
		QuotaSlotsPerDeployment: int(valOrDefault(req.Msg.SlotsPerDeployment, uint32(org.QuotaSlotsPerDeployment))),
		QuotaOutstandingInvites: int(valOrDefault(req.Msg.OutstandingInvites, uint32(org.QuotaOutstandingInvites))),
	}

	updatedOrg, err := s.admin.DB.UpdateOrganization(ctx, org.ID, opts)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&adminv1.SudoUpdateOrganizationQuotasResponse{
		Organization: organizationToDTO(updatedOrg),
	}), nil
}

func organizationToDTO(o *database.Organization) *adminv1.Organization {
	return &adminv1.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		Quotas: &adminv1.OrganizationQuotas{
			Projects:           uint32(o.QuotaProjects),
			Deployments:        uint32(o.QuotaDeployments),
			SlotsTotal:         uint32(o.QuotaSlotsTotal),
			SlotsPerDeployment: uint32(o.QuotaSlotsPerDeployment),
			OutstandingInvites: uint32(o.QuotaOutstandingInvites),
		},
		CreatedOn: timestamppb.New(o.CreatedOn),
		UpdatedOn: timestamppb.New(o.UpdatedOn),
	}
}

func whitelistedDomainToPB(a *database.OrganizationWhitelistedDomainWithJoinedRoleNames) *adminv1.WhitelistedDomain {
	return &adminv1.WhitelistedDomain{
		Domain: a.Domain,
		Role:   a.RoleName,
	}
}
