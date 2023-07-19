package server

import (
	"context"

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

	return &adminv1.CreateServiceResponse{Service: &adminv1.Service{}}, nil
}

// ListServices lists all service accounts.
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

	return &adminv1.ListServicesResponse{}, nil
}

// UpdateService updates a service account.
func (s *Server) UpdateService(ctx context.Context, req *adminv1.UpdateServiceRequest) (*adminv1.UpdateServiceResponse, error) {
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

	return &adminv1.UpdateServiceResponse{}, nil
}

// DeleteService deletes a service account.
func (s *Server) DeleteService(ctx context.Context, req *adminv1.DeleteServiceRequest) (*adminv1.DeleteServiceResponse, error) {
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

	return &adminv1.DeleteServiceResponse{}, nil
}
