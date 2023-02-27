package server

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin/database"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// FindOrganizations implements AdminService.
// (GET /v1/organizations)
func (s *Server) FindOrganizations(ctx context.Context, req *adminv1.FindOrganizationsRequest) (*adminv1.FindOrganizationsResponse, error) {
	orgs, err := s.db.FindOrganizations(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs := make([]*adminv1.Organization, len(orgs))
	for i, org := range orgs {
		pbs[i] = orgToDTO(org)
	}

	return &adminv1.FindOrganizationsResponse{Organization: pbs}, nil
}

// (GET /organizations/{name})
func (s *Server) FindOrganization(ctx context.Context, req *adminv1.FindOrganizationRequest) (*adminv1.FindOrganizationResponse, error) {
	org, err := s.db.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "org not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.FindOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

// CreateOrganization implements AdminService.
// (POST /organizations)
func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	org, err := s.db.CreateOrganization(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

// (DELETE /organizations/{name})
func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	err := s.db.DeleteOrganization(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteOrganizationResponse{
		Name: req.Name,
	}, nil
}

// (PUT /organizations/{name})
func (s *Server) UpdateOrganization(ctx context.Context, req *adminv1.UpdateOrganizationRequest) (*adminv1.UpdateOrganizationResponse, error) {
	org, err := s.db.UpdateOrganization(ctx, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.UpdateOrganizationResponse{
		Organization: orgToDTO(org),
	}, nil
}

func orgToDTO(o *database.Organization) *adminv1.Organization {
	return &adminv1.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		CreatedOn:   timestamppb.New(o.CreatedOn),
		UpdatedOn:   timestamppb.New(o.CreatedOn),
	}
}
