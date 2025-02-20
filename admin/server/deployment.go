package server

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: See details in api.proto.
func (s *Server) TriggerReconcile(ctx context.Context, req *adminv1.TriggerReconcileRequest) (*adminv1.TriggerReconcileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	err = s.admin.TriggerParser(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerReconcileResponse{}, nil
}

// Deprecated: See details in api.proto.
func (s *Server) TriggerRefreshSources(ctx context.Context, req *adminv1.TriggerRefreshSourcesRequest) (*adminv1.TriggerRefreshSourcesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
		attribute.StringSlice("args.sources", req.Sources),
	)

	depl, err := s.admin.DB.FindDeployment(ctx, req.DeploymentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	var names []*runtimev1.ResourceName
	for _, source := range req.Sources {
		names = append(names, &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: source})
	}

	rt, err := s.admin.OpenRuntimeClient(depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId:       depl.RuntimeInstanceID,
		Resources:        names,
		AllSourcesModels: len(names) == 0, // Backwards compatibility
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRefreshSourcesResponse{}, nil
}

// GetDeploymentCredentials returns runtime info and JWT on behalf of a specific user, or alternatively for a raw set of JWT attributes
func (s *Server) GetDeploymentCredentials(ctx context.Context, req *adminv1.GetDeploymentCredentialsRequest) (*adminv1.GetDeploymentCredentialsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.ttl_seconds", strconv.FormatUint(uint64(req.TtlSeconds), 10)),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment")
	}

	prodDepl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Branch != "" && req.Branch != prodDepl.Branch {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment for given branch")
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	// If the user is not a superuser, they must have ManageProd permissions
	if !permissions.ManageProd && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	var attr map[string]any
	if req.For != nil {
		switch forVal := req.For.(type) {
		case *adminv1.GetDeploymentCredentialsRequest_UserId:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, forVal.UserId, "")
			if err != nil {
				return nil, err
			}
		case *adminv1.GetDeploymentCredentialsRequest_UserEmail:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, "", forVal.UserEmail)
			if err != nil {
				return nil, err
			}
		case *adminv1.GetDeploymentCredentialsRequest_Attributes:
			attr = forVal.Attributes.AsMap()
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
		}
	}

	ttlDuration := runtimeAccessTokenEmbedTTL
	if req.TtlSeconds > 0 {
		ttlDuration = time.Duration(req.TtlSeconds) * time.Second
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: prodDepl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         ttlDuration,
		InstancePermissions: map[string][]runtimeauth.Permission{
			prodDepl.RuntimeInstanceID: {
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadAPI,
			},
		},
		Attributes: attr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(prodDepl.ID)

	return &adminv1.GetDeploymentCredentialsResponse{
		RuntimeHost: prodDepl.RuntimeHost,
		InstanceId:  prodDepl.RuntimeInstanceID,
		AccessToken: jwt,
		TtlSeconds:  uint32(ttlDuration.Seconds()),
	}, nil
}

func (s *Server) GetIFrame(ctx context.Context, req *adminv1.GetIFrameRequest) (*adminv1.GetIFrameResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.type", req.Type),
		attribute.String("args.kind", req.Kind), // nolint:staticcheck // Deprecated but still used
		attribute.String("args.resource", req.Resource),
		attribute.String("args.ttl_seconds", strconv.FormatUint(uint64(req.TtlSeconds), 10)),
		attribute.String("args.state", req.State),
	)

	if !req.Navigation && req.Resource == "" {
		return nil, status.Error(codes.InvalidArgument, "resource must be provided if navigation is not enabled")
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if proj.ProdDeploymentID == nil {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment")
	}

	prodDepl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Branch != "" && req.Branch != prodDepl.Branch {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment for given branch")
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	// If the user is not a superuser, they must have ManageProd permissions
	if !permissions.ManageProd && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	var attr map[string]any
	if req.For != nil {
		switch forVal := req.For.(type) {
		case *adminv1.GetIFrameRequest_UserId:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, forVal.UserId, "")
			if err != nil {
				return nil, err
			}
		case *adminv1.GetIFrameRequest_UserEmail:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, "", forVal.UserEmail)
			if err != nil {
				return nil, err
			}
		case *adminv1.GetIFrameRequest_Attributes:
			attr = forVal.Attributes.AsMap()
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
		}
	}

	ttlDuration := runtimeAccessTokenEmbedTTL
	if req.TtlSeconds > 0 {
		ttlDuration = time.Duration(req.TtlSeconds) * time.Second
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: prodDepl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         ttlDuration,
		InstancePermissions: map[string][]runtimeauth.Permission{
			prodDepl.RuntimeInstanceID: {
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadAPI,
			},
		},
		Attributes: attr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(prodDepl.ID)

	iframeQuery := map[string]string{
		"runtime_host": prodDepl.RuntimeHost,
		"instance_id":  prodDepl.RuntimeInstanceID,
		"access_token": jwt,
	}

	if req.Type != "" {
		iframeQuery["type"] = req.Type
	} else if req.Kind != "" { // nolint:staticcheck // Deprecated but still used
		iframeQuery["type"] = req.Kind
	} else {
		// Default to an explore if no type is explicitly provided
		iframeQuery["type"] = runtime.ResourceKindExplore
	}
	iframeQuery["kind"] = iframeQuery["type"] // For backwards compatibility

	if req.Resource != "" {
		iframeQuery["resource"] = req.Resource
	}

	if req.Theme != "" {
		iframeQuery["theme"] = req.Theme
	}

	if req.Navigation {
		iframeQuery["navigation"] = "true"
	}

	if req.State != "" {
		iframeQuery["state"] = req.State
	}

	for k, v := range req.Query {
		iframeQuery[k] = v
	}

	iFrameURL, err := s.admin.URLs.Embed(iframeQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not construct iframe url: %s", err.Error())
	}

	return &adminv1.GetIFrameResponse{
		IframeSrc:   iFrameURL,
		RuntimeHost: prodDepl.RuntimeHost,
		InstanceId:  prodDepl.RuntimeInstanceID,
		AccessToken: jwt,
		TtlSeconds:  uint32(ttlDuration.Seconds()),
	}, nil
}

// getAttributesFor returns a map of attributes for a given user and project.
// The caller should only provide one of userID or userEmail (if both or neither are set, an error will be returned).
// NOTE: The value returned from this function must be valid for structpb.NewStruct (e.g. must use []any for slices, not a more specific slice type).
func (s *Server) getAttributesForUser(ctx context.Context, orgID, projID, userID, userEmail string) (map[string]any, error) {
	if userID == "" && userEmail == "" {
		return nil, errors.New("must provide either userID or userEmail")
	}

	if userEmail != "" {
		if userID != "" {
			return nil, errors.New("must provide either userID or userEmail, not both")
		}

		user, err := s.admin.DB.FindUserByEmail(ctx, userEmail)
		if err != nil {
			// For user attributes, we do not require the email to exist as a Rill user.
			// For example, the attributes may be used for a dashboard embedded as an iframe on a third-party website.
			// For these cases, we return attributes that present the email as a non-admin user.
			if errors.Is(err, database.ErrNotFound) {
				return map[string]any{
					"email":  userEmail,
					"domain": userEmail[strings.LastIndex(userEmail, "@")+1:],
					"admin":  false,
				}, nil
			}
			return nil, err
		}

		userID = user.ID
	}

	forOrgPerms, err := s.admin.OrganizationPermissionsForUser(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	forProjPerms, err := s.admin.ProjectPermissionsForUser(ctx, projID, userID, forOrgPerms)
	if err != nil {
		return nil, err
	}

	attr, err := s.jwtAttributesForUser(ctx, userID, orgID, forProjPerms)
	if err != nil {
		return nil, err
	}

	return attr, nil
}
