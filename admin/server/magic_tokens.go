package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const magicAuthTokenFilterMaxSize = 1024

// magicAuthTokenSessionTTL is the validity of a session issued by UnlockMagicAuthToken.
// A session also expires when the magic auth token it belongs to expires.
const magicAuthTokenSessionTTL = 24 * time.Hour

func (s *Server) IssueMagicAuthToken(ctx context.Context, req *adminv1.IssueMagicAuthTokenRequest) (*adminv1.IssueMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.display_name", req.DisplayName),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.CreateMagicAuthTokens {
		return nil, status.Error(codes.PermissionDenied, "not allowed to create a magic auth token")
	}

	if req.ResourceName != "" && req.ResourceType != "" { // nolint:staticcheck // for backwards compatibility
		addResource := true
		for _, r := range req.Resources {
			if r.Type == req.ResourceType && r.Name == req.ResourceName { // nolint:staticcheck // for backwards compatibility
				addResource = false
				break
			}
		}
		if addResource {
			req.Resources = append(req.Resources, &adminv1.ResourceName{
				Type: req.ResourceType, // nolint:staticcheck // for backwards compatibility
				Name: req.ResourceName, // nolint:staticcheck // for backwards compatibility
			})
		}
	}

	resources := make([]database.ResourceName, len(req.Resources))
	for i, r := range req.Resources {
		resources[i] = database.ResourceName{
			Type: r.Type,
			Name: r.Name,
		}
	}
	opts := &admin.IssueMagicAuthTokenOptions{
		ProjectID:   proj.ID,
		Fields:      req.Fields,
		State:       req.State,
		DisplayName: req.DisplayName,
		Resources:   resources,
	}

	if req.TtlMinutes != 0 {
		ttl := time.Duration(req.TtlMinutes) * time.Minute
		opts.TTL = &ttl
	}

	if req.Password != "" {
		err := validateMagicAuthTokenPassword(req.Password)
		if err != nil {
			return nil, err
		}
		opts.Password = req.Password
	}

	if claims.OwnerType() == auth.OwnerTypeUser {
		id := claims.OwnerID()
		opts.CreatedByUserID = &id

		// Generate JWT attributes based on the creating user's, but with limited project-level permissions.
		// We store these attributes with the magic token, so it can simulate the creating user (even if the creating user is later deleted or their permissions change).
		//
		// NOTE: A problem with this approach is that if we change the built-in format of JWT attributes, these will remain as they were when captured.
		// NOTE: Another problem is that if the creator is an admin, attrs["admin"] will be true. It shouldn't be a problem today, but could end up leaking some privileges in the future if we're not careful.
		attrs, err := s.jwtAttributesForUser(ctx, claims.OwnerID(), proj.OrganizationID, projPerms)
		if err != nil {
			return nil, err
		}
		opts.Attributes = attrs
	}

	filterSize := 0
	for mv, filter := range req.MetricsViewFilters {
		if opts.MetricsViewFilterJSONs == nil {
			opts.MetricsViewFilterJSONs = make(map[string]string)
		}
		val, err := protojson.Marshal(filter)
		if err != nil {
			return nil, err
		}

		filterSize += len(val)
		if filterSize > magicAuthTokenFilterMaxSize {
			return nil, status.Errorf(codes.InvalidArgument, "filter size exceeds limit (got %d bytes, but the limit is %d bytes)", len(val), magicAuthTokenFilterMaxSize)
		}

		filterJSON := string(val)

		opts.MetricsViewFilterJSONs[mv] = filterJSON
	}

	token, err := s.admin.IssueMagicAuthToken(ctx, opts)
	if err != nil {
		return nil, err
	}

	tokenStr := token.Token().String()
	return &adminv1.IssueMagicAuthTokenResponse{
		Token: tokenStr,
		Url:   s.admin.URLs.WithCustomDomain(org.CustomDomain).MagicAuthTokenOpen(req.Org, req.Project, tokenStr),
	}, nil
}

func (s *Server) GetCurrentMagicAuthToken(ctx context.Context, req *adminv1.GetCurrentMagicAuthTokenRequest) (*adminv1.GetCurrentMagicAuthTokenResponse, error) {
	claims := auth.GetClaims(ctx)
	switch claims.OwnerType() {
	case auth.OwnerTypeMagicAuthToken:
	case auth.OwnerTypeLockedMagicAuthToken:
		// Until the token is unlocked with its password, reveal nothing but the fact that a password is required.
		return &adminv1.GetCurrentMagicAuthTokenResponse{
			Token: &adminv1.MagicAuthToken{
				Id:                claims.OwnerID(),
				PasswordProtected: true,
			},
			Locked: true,
		}, nil
	default:
		return nil, status.Error(codes.PermissionDenied, "request was not made with a magic auth token")
	}

	tkn, err := s.admin.DB.FindMagicAuthTokenWithUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("project with id %s not found", tkn.ProjectID))
		}
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	pb, err := s.magicAuthTokenToPB(tkn, org, proj)
	if err != nil {
		return nil, err
	}

	return &adminv1.GetCurrentMagicAuthTokenResponse{
		Token: pb,
	}, nil
}

func (s *Server) ListMagicAuthTokens(ctx context.Context, req *adminv1.ListMagicAuthTokensRequest) (*adminv1.ListMagicAuthTokensResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.CreateMagicAuthTokens && !projPerms.ManageMagicAuthTokens {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	var createdByUserID *string
	if !projPerms.ManageMagicAuthTokens {
		if claims.OwnerType() != auth.OwnerTypeUser {
			return nil, status.Error(codes.PermissionDenied, "not allowed to manage magic auth tokens")
		}

		id := claims.OwnerID()
		createdByUserID = &id
	}

	tokens, err := s.admin.DB.FindMagicAuthTokensWithUser(ctx, proj.ID, createdByUserID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if len(tokens) >= pageSize {
		nextPageToken = marshalPageToken(tokens[len(tokens)-1].ID)
	}

	pbs, err := s.magicAuthTokensToPB(tokens, org, proj)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.ListMagicAuthTokensResponse{
		Tokens:        pbs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) RevokeMagicAuthToken(ctx context.Context, req *adminv1.RevokeMagicAuthTokenRequest) (*adminv1.RevokeMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	tkn, err := s.admin.DB.FindMagicAuthToken(ctx, req.TokenId, false)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.ManageMagicAuthTokens {
		// If they don't have manage permissions, they can only revoke tokens they created themselves.
		isCreator := tkn.CreatedByUserID != nil && *tkn.CreatedByUserID == claims.OwnerID()
		if !projPerms.CreateMagicAuthTokens || !isCreator {
			return nil, status.Error(codes.PermissionDenied, "not allowed to revoke this magic auth token")
		}
	}

	err = s.admin.DB.DeleteMagicAuthToken(ctx, tkn.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RevokeMagicAuthTokenResponse{}, nil
}

func (s *Server) UpdateMagicAuthToken(ctx context.Context, req *adminv1.UpdateMagicAuthTokenRequest) (*adminv1.UpdateMagicAuthTokenResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.token_id", req.TokenId),
	)

	if req.Password != nil && req.RemovePassword {
		return nil, status.Error(codes.InvalidArgument, "password and remove_password can't both be set")
	}

	tkn, err := s.admin.DB.FindMagicAuthToken(ctx, req.TokenId, false)
	if err != nil {
		return nil, err
	}

	proj, err := s.admin.DB.FindProject(ctx, tkn.ProjectID)
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	projPerms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !projPerms.ManageMagicAuthTokens {
		// If they don't have manage permissions, they can only update tokens they created themselves.
		isCreator := tkn.CreatedByUserID != nil && *tkn.CreatedByUserID == claims.OwnerID()
		if !projPerms.CreateMagicAuthTokens || !isCreator {
			return nil, status.Error(codes.PermissionDenied, "not allowed to update this magic auth token")
		}
	}

	opts := &database.UpdateMagicAuthTokenOptions{
		DisplayName:  tkn.DisplayName,
		PasswordHash: tkn.PasswordHash,
	}
	if req.DisplayName != nil {
		opts.DisplayName = *req.DisplayName
	}
	passwordChanged := false
	if req.Password != nil {
		err := validateMagicAuthTokenPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		opts.PasswordHash, err = admin.HashMagicAuthTokenPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		passwordChanged = true
	} else if req.RemovePassword {
		opts.PasswordHash = nil
		passwordChanged = len(tkn.PasswordHash) > 0
	}

	txCtx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = s.admin.DB.UpdateMagicAuthToken(txCtx, tkn.ID, opts)
	if err != nil {
		return nil, err
	}

	// Recipients who unlocked the token with the old password must enter the new one.
	if passwordChanged {
		err = s.admin.DB.DeleteMagicAuthTokenSessionsForToken(txCtx, tkn.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	updated, err := s.admin.DB.FindMagicAuthTokenWithUser(ctx, tkn.ID)
	if err != nil {
		return nil, err
	}

	pb, err := s.magicAuthTokenToPB(updated, org, proj)
	if err != nil {
		return nil, err
	}

	return &adminv1.UpdateMagicAuthTokenResponse{
		Token: pb,
	}, nil
}

func (s *Server) UnlockMagicAuthToken(ctx context.Context, req *adminv1.UnlockMagicAuthTokenRequest) (*adminv1.UnlockMagicAuthTokenResponse, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeLockedMagicAuthToken {
		return nil, status.Error(codes.PermissionDenied, "request was not made with a locked magic auth token")
	}

	// The claims are derived from a token model that may come from the auth cache, so the password is not checked against them:
	// a password that was just changed must not unlock the token until the cache catches up.
	sess, sessTkn, err := s.admin.UnlockMagicAuthToken(ctx, claims.OwnerID(), req.Password, magicAuthTokenSessionTTL)
	if err != nil {
		if errors.Is(err, admin.ErrMagicAuthTokenPasswordIncorrect) {
			return nil, status.Error(codes.PermissionDenied, "incorrect password")
		}
		if errors.Is(err, admin.ErrMagicAuthTokenNotPasswordProtected) {
			return nil, status.Error(codes.FailedPrecondition, "the public URL is no longer password protected")
		}
		return nil, err
	}

	return &adminv1.UnlockMagicAuthTokenResponse{
		SessionToken: sessTkn.String(),
		ExpiresOn:    timestamppb.New(sess.ExpiresOn),
	}, nil
}

func (s *Server) magicAuthTokensToPB(tkns []*database.MagicAuthTokenWithUser, org *database.Organization, proj *database.Project) ([]*adminv1.MagicAuthToken, error) {
	var pbs []*adminv1.MagicAuthToken
	for _, tkn := range tkns {
		pb, err := s.magicAuthTokenToPB(tkn, org, proj)
		if err != nil {
			return nil, err
		}
		pbs = append(pbs, pb)
	}
	return pbs, nil
}

func (s *Server) magicAuthTokenToPB(tkn *database.MagicAuthTokenWithUser, org *database.Organization, proj *database.Project) (*adminv1.MagicAuthToken, error) {
	attrs, err := structpb.NewStruct(tkn.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert attributes to structpb: %w", err)
	}

	mvFilters := make(map[string]*runtimev1.Expression)
	for mv, filterJSON := range tkn.MetricsViewFilterJSONs {
		mvFilter := &runtimev1.Expression{}
		err := protojson.Unmarshal([]byte(filterJSON), mvFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter for metrics view %s: %w", mv, err)
		}
		mvFilters[mv] = mvFilter
	}

	// backwards compatibility
	tokenStr := ""
	url := ""
	if len(tkn.Secret) != 0 {
		id, err := uuid.Parse(tkn.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token ID: %w", err)
		}
		token, err := authtoken.FromParts(authtoken.TypeMagic, id, tkn.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to create magic auth token from parts: %w", err)
		}
		tokenStr = token.String()
		url = s.admin.URLs.WithCustomDomain(org.CustomDomain).MagicAuthTokenOpen(org.Name, proj.Name, tokenStr)
	}

	rs := make([]*adminv1.ResourceName, len(tkn.Resources))
	for i, r := range tkn.Resources {
		rs[i] = &adminv1.ResourceName{
			Type: r.Type,
			Name: r.Name,
		}
	}

	res := &adminv1.MagicAuthToken{
		Id:                 tkn.ID,
		ProjectId:          tkn.ProjectID,
		Url:                url,
		Token:              tokenStr,
		CreatedOn:          timestamppb.New(tkn.CreatedOn),
		ExpiresOn:          nil,
		UsedOn:             timestamppb.New(tkn.UsedOn),
		CreatedByUserId:    safeStr(tkn.CreatedByUserID),
		CreatedByUserEmail: tkn.CreatedByUserEmail,
		Attributes:         attrs,
		Resources:          rs,
		MetricsViewFilters: mvFilters,
		Fields:             tkn.Fields,
		State:              tkn.State,
		DisplayName:        tkn.DisplayName,
		PasswordProtected:  len(tkn.PasswordHash) > 0,
	}
	if tkn.ExpiresOn != nil {
		res.ExpiresOn = timestamppb.New(*tkn.ExpiresOn)
	}
	return res, nil
}

// validateMagicAuthTokenPassword checks the length constraints of a password for a magic auth token.
func validateMagicAuthTokenPassword(password string) error {
	if err := admin.ValidateMagicAuthTokenPassword(password); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
