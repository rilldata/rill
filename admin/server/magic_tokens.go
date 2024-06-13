package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) IssueMagicAuthToken(ctx context.Context, req *adminv1.IssueMagicAuthTokenRequest) (*adminv1.IssueMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.dashboard", req.Dashboard),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("project %q not found", req.Project))
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to create a magic auth token")
	}

	var ttl *time.Duration
	if req.TtlMinutes != 0 {
		ttlVal := time.Duration(req.TtlMinutes) * time.Minute
		ttl = &ttlVal
	}

	var filterJSON string
	if req.PresetFilter != nil {
		res, err := protojson.Marshal(req.PresetFilter)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		filterJSON = string(res)
	}

	token, err := s.admin.IssueMagicAuthToken(ctx, proj.ID, ttl, req.Dashboard, filterJSON, req.ExcludeFields)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.IssueMagicAuthTokenResponse{
		Token: token.Token().String(),
	}, nil
}

func (s *Server) ListMagicAuthTokens(ctx context.Context, req *adminv1.ListMagicAuthTokensRequest) (*adminv1.ListMagicAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("project %q not found", req.Project))
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
	}

	tokens, err := s.admin.DB.FindMagicAuthTokens(ctx, proj.ID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(tokens) >= pageSize {
		nextToken = marshalPageToken(tokens[len(tokens)-1].ID)
	}

	return &adminv1.ListMagicAuthTokensResponse{
		Tokens:        magicAuthTokensToPB(tokens),
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) RevokeMagicAuthToken(ctx context.Context, req *adminv1.RevokeMagicAuthTokenRequest) (*adminv1.RevokeMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	tkn, err := s.admin.DB.FindMagicAuthToken(ctx, req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find project for token: %v", err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
	}

	err = s.admin.DB.DeleteMagicAuthToken(ctx, tkn.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.RevokeMagicAuthTokenResponse{}, nil
}

func magicAuthTokenToPB(tkn *database.MagicAuthToken) *adminv1.MagicAuthToken {
	return &adminv1.MagicAuthToken{
		Id: tkn.ID,

		// TODO:

		CreatedOn: timestamppb.New(tkn.CreatedOn),
		ExpiresOn: timestamppb.New(safeTime(tkn.ExpiresOn)),
	}
}

func magicAuthTokensToPB(tkns []*database.MagicAuthToken) []*adminv1.MagicAuthToken {
	var pbs []*adminv1.MagicAuthToken
	for _, tkn := range tkns {
		pbs = append(pbs, magicAuthTokenToPB(tkn))
	}
	return pbs
}
