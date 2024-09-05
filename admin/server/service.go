package server

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateService creates a new service account.
func (s *Server) CreateService(ctx context.Context, req *adminv1.CreateServiceRequest) (*adminv1.CreateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization_name", req.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to create a service")
	}

	service, err := s.admin.DB.InsertService(ctx, &database.InsertServiceOptions{
		OrgID: org.ID,
		Name:  req.Name,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateServiceResponse{
		Service: serviceToPB(service, req.OrganizationName),
	}, nil
}

// ListServices lists all service accounts in an organization.
func (s *Server) ListServices(ctx context.Context, req *adminv1.ListServicesRequest) (*adminv1.ListServicesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.OrganizationName),
		attribute.String("args.O]organization_name", req.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	// Need a check if any other service permission to list a services
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to list services")
	}

	services, err := s.admin.DB.FindServicesByOrgID(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ListServicesResponse{
		Services: servicesToPB(services, req.OrganizationName),
	}, nil
}

// UpdateService updates a service account.
func (s *Server) UpdateService(ctx context.Context, req *adminv1.UpdateServiceRequest) (*adminv1.UpdateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization_name", req.OrganizationName),
	)

	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update service")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedService, err := s.admin.DB.UpdateService(ctx, service.ID, &database.UpdateServiceOptions{
		Name: valOrDefault(req.NewName, service.Name),
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateServiceResponse{
		Service: serviceToPB(updatedService, req.OrganizationName),
	}, nil
}

// DeleteService deletes a service account.
func (s *Server) DeleteService(ctx context.Context, req *adminv1.DeleteServiceRequest) (*adminv1.DeleteServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.organization_name", req.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete service")
	}

	err = s.admin.DB.DeleteService(ctx, service.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.DeleteServiceResponse{}, nil
}

// ListServiceAuthTokens lists all auth tokens for a service account.
func (s *Server) ListServiceAuthTokens(ctx context.Context, req *adminv1.ListServiceAuthTokensRequest) (*adminv1.ListServiceAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.service_name", req.ServiceName),
		attribute.String("args.organization_name", req.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.ServiceName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokens, err := s.admin.DB.FindServiceAuthTokens(ctx, service.ID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*adminv1.ServiceToken, len(tokens))
	for i, token := range tokens {
		dtos[i] = &adminv1.ServiceToken{
			Id:        token.ID,
			CreatedOn: timestamppb.New(token.CreatedOn),
			ExpiresOn: timestamppb.New(safeTime(token.ExpiresOn)),
		}
	}

	return &adminv1.ListServiceAuthTokensResponse{
		Tokens: dtos,
	}, nil
}

// IssueServiceAuthToken issues a new auth token for a service account.
func (s *Server) IssueServiceAuthToken(ctx context.Context, req *adminv1.IssueServiceAuthTokenRequest) (*adminv1.IssueServiceAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.service_name", req.ServiceName),
		attribute.String("args.organization_name", req.OrganizationName),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	service, err := s.admin.DB.FindServiceByName(ctx, org.ID, req.ServiceName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	token, err := s.admin.IssueServiceAuthToken(ctx, service.ID, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.IssueServiceAuthTokenResponse{
		Token: token.Token().String(),
	}, nil
}

// RevokServiceAuthToken revokes an auth token for a service account.
func (s *Server) RevokeServiceAuthToken(ctx context.Context, req *adminv1.RevokeServiceAuthTokenRequest) (*adminv1.RevokeServiceAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	token, err := s.admin.DB.FindServiceAuthToken(ctx, req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	service, err := s.admin.DB.FindService(ctx, token.ServiceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, service.OrgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to revoke auth token")
	}

	err = s.admin.DB.DeleteServiceAuthToken(ctx, token.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RevokeServiceAuthTokenResponse{}, nil
}

func serviceToPB(service *database.Service, orgName string) *adminv1.Service {
	return &adminv1.Service{
		Id:        service.ID,
		Name:      service.Name,
		OrgId:     service.OrgID,
		OrgName:   orgName,
		CreatedOn: timestamppb.New(service.CreatedOn),
		UpdatedOn: timestamppb.New(service.UpdatedOn),
	}
}

func servicesToPB(services []*database.Service, orgName string) []*adminv1.Service {
	var pbServices []*adminv1.Service
	for _, service := range services {
		pbServices = append(pbServices, serviceToPB(service, orgName))
	}
	return pbServices
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
