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
		attribute.String("args.org", req.OrganizationName),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	return &adminv1.CreateServiceResponse{Service: &adminv1.Service{}}, nil
}

// func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
// 	observability.AddRequestAttributes(ctx,
// 		attribute.String("args.org", req.Name),
// 		attribute.String("args.description", req.Description),
// 	)

// 	// Check the request is made by an authenticated user
// 	claims := auth.GetClaims(ctx)
// 	if claims.OwnerType() != auth.OwnerTypeUser {
// 		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
// 	}

// 	// check single user org limit for this user
// 	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}
// 	count, err := s.admin.DB.CountSingleuserOrganizationsForMemberUser(ctx, user.ID)
// 	if err != nil {
// 		return nil, status.Error(codes.Internal, err.Error())
// 	}
// 	if user.QuotaSingleuserOrgs >= 0 && count >= user.QuotaSingleuserOrgs {
// 		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: you can only create %d single-user orgs", user.QuotaSingleuserOrgs)
// 	}

// 	org, err := s.admin.CreateOrganizationForUser(ctx, user.ID, req.Name, req.Description)
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, err.Error())
// 	}

// 	return &adminv1.CreateOrganizationResponse{
// 		Organization: organizationToDTO(org),
// 	}, nil
// }
