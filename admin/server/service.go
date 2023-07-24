package server

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateService creates a new service account.
func (s *Server) CreateService(ctx context.Context, req *adminv1.CreateServiceRequest) (*adminv1.CreateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.OrganizationName", req.OrganizationName),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Need a check if any other service permission to create a service
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

	service, err := s.admin.CreateServiceForOrganization(ctx, org.Name, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.CreateServiceResponse{
		Service: serviceToPB(service),
	}, nil
}

// ListServices lists all service accounts in an organization.
func (s *Server) ListServices(ctx context.Context, req *adminv1.ListServicesRequest) (*adminv1.ListServicesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.OrganizationName),
		attribute.String("args.OrganizationName", req.OrganizationName),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Need a check if any other service permission to list a services
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

	services, err := s.admin.DB.FindServicesByOrgName(ctx, org.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ListServicesResponse{
		Services: servicesToPB(services),
	}, nil
}

// UpdateService updates a service account.
func (s *Server) UpdateService(ctx context.Context, req *adminv1.UpdateServiceRequest) (*adminv1.UpdateServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.OrganizationName", req.OrganizationName),
	)

	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Need a check for service permission to update a service
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

	service, err = s.admin.DB.UpdateService(ctx, service.ID, &database.UpdateServiceOptions{
		Name: valOrDefault(req.NewName, service.Name),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UpdateServiceResponse{
		Service: serviceToPB(service),
	}, nil
}

// DeleteService deletes a service account.
func (s *Server) DeleteService(ctx context.Context, req *adminv1.DeleteServiceRequest) (*adminv1.DeleteServiceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.name", req.Name),
		attribute.String("args.OrganizationName", req.OrganizationName),
	)

	service, err := s.admin.DB.FindServiceByName(ctx, req.OrganizationName, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Need a check for service permission to delete a service
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

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

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Need a check for service permission to list a service auth tokens
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

	service, err := s.admin.DB.FindServiceByName(ctx, req.OrganizationName, req.ServiceName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokens, err := s.admin.ListServiceAuthTokens(ctx, service.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	dtos := make([]string, len(tokens))
	for i, p := range tokens {
		dtos[i] = p.Token().String()
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

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	service, err := s.admin.DB.FindServiceByName(ctx, req.OrganizationName, req.ServiceName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Need a check for service permission to issue a token
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

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

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	// this is not required
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Need a check for service permission to revoke a token
	// if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
	// 	return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	// }

	err := s.admin.RevokeAuthToken(ctx, req.TokenId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RevokeServiceAuthTokenResponse{}, nil
}

func serviceToPB(service *database.Service) *adminv1.Service {
	return &adminv1.Service{
		Id:          service.ID,
		ServiceName: service.Name,
		OrgName:     service.OrgName,
	}
}

func servicesToPB(services []*database.Service) []*adminv1.Service {
	var pbServices []*adminv1.Service
	for _, service := range services {
		pbServices = append(pbServices, serviceToPB(service))
	}
	return pbServices
}
