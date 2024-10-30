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

func (s *Server) SudoIssueRuntimeManagerToken(ctx context.Context, req *adminv1.SudoIssueRuntimeManagerTokenRequest) (*adminv1.SudoIssueRuntimeManagerTokenResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.host", req.Host))

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can request a manager token")
	}

	// In production, the REST and gRPC endpoints are the same, but in development, they're served on different ports.
	// And the runtime audience uses the REST endpoint.
	// TODO: Move to HTTP and gRPC on the same host for local development.
	// Until we make that change, this is a convenient hack for local development (assumes REST on port 8081).
	if req.Host == "http://localhost:9091" {
		req.Host = "http://localhost:8081"
	}

	jwt, err := s.admin.IssueRuntimeManagementToken(req.Host)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to issue runtime manager token: %v", err)
	}

	return &adminv1.SudoIssueRuntimeManagerTokenResponse{
		Token: jwt,
	}, nil
}
