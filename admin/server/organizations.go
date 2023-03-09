package server

import (
	"context"
	"errors"

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

	orgs, err := s.admin.DB.FindOrganizations(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// filter out orgs that the user doesn't have access to
	filteredOrgs := make([]*database.Organization, 0, len(orgs))
	for _, org := range orgs {
		if claims.CanOrganization(ctx, org.ID, auth.ReadOrg) {
			filteredOrgs = append(filteredOrgs, org)
		}
	}

	pbs := make([]*adminv1.Organization, len(filteredOrgs))
	for i, org := range filteredOrgs {
		pbs[i] = orgToDTO(org)
	}

	return &adminv1.ListOrganizationsResponse{Organizations: pbs}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *adminv1.GetOrganizationRequest) (*adminv1.GetOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.CanOrganization(ctx, org.ID, auth.ReadOrg) {
		return nil, status.Error(codes.PermissionDenied, "not allowed to read org")
	}

	return &adminv1.GetOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.InsertOrganization(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

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
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

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

	org, err = s.admin.DB.UpdateOrganization(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UpdateOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

func (s *Server) ListOrgMembers(ctx context.Context, req *adminv1.ListOrgMembersRequest) (*adminv1.ListOrgMembersResponse, error) {
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

	users, err := s.admin.DB.FindOrganizationMembers(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dtos := make([]*adminv1.User, len(users))
	for i, user := range users {
		dtos[i] = userToPB(user)
	}

	return &adminv1.ListOrgMembersResponse{Users: dtos}, nil
}

func (s *Server) AddOrgMember(ctx context.Context, req *adminv1.AddOrgMemberRequest) (*adminv1.AddOrgMemberResponse, error) {
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
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org members")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.admin.DB.AddOrganizationMember(ctx, org.ID, user.ID, role.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddOrgMemberResponse{}, nil
}

func (s *Server) RemoveOrgMember(ctx context.Context, req *adminv1.RemoveOrgMemberRequest) (*adminv1.RemoveOrgMemberResponse, error) {
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

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMembersByRole(ctx, org.ID, database.RoleIDOrgAdmin)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(users) == 1 && users[0].Email == req.Email {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	err = s.admin.DB.RemoveOrganizationMember(ctx, org.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.RemoveOrgMemberResponse{}, nil
}

func (s *Server) SetOrgMemberRole(ctx context.Context, req *adminv1.SetOrgMemberRoleRequest) (*adminv1.SetOrgMemberRoleResponse, error) {
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
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	err = s.admin.DB.UpdateOrganizationMemberRole(ctx, org.ID, req.Email, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.SetOrgMemberRoleResponse{}, nil
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

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMembersByRole(ctx, org.ID, database.RoleIDOrgAdmin)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(users) == 1 && users[0].ID == claims.OwnerID() {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	err = s.admin.DB.RemoveOrganizationMember(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.LeaveOrganizationResponse{}, nil
}

func orgToDTO(o *database.Organization) *adminv1.Organization {
	return &adminv1.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		CreatedOn:   timestamppb.New(o.CreatedOn),
		UpdatedOn:   timestamppb.New(o.UpdatedOn),
	}
}
