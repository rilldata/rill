package server

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetIFrame(ctx context.Context, req *adminv1.GetIFrameRequest) (*adminv1.GetIFrameResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.kind", req.Kind),
		attribute.String("args.resource", req.Resource),
		attribute.String("args.ttl_seconds", strconv.FormatUint(uint64(req.TtlSeconds), 10)),
		attribute.String("args.state", req.State),
	)

	if req.Resource == "" {
		return nil, status.Error(codes.InvalidArgument, "resource must be specified")
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
	switch forVal := req.For.(type) {
	case *adminv1.GetIFrameRequest_UserId:
		attr, err = s.getAttributesFor(ctx, forVal.UserId, proj.OrganizationID, proj.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	case *adminv1.GetIFrameRequest_UserEmail:
		user, err := s.admin.DB.FindUserByEmail(ctx, forVal.UserEmail)
		if errors.Is(err, database.ErrNotFound) {
			attr = map[string]any{
				"email":  forVal.UserEmail,
				"domain": forVal.UserEmail[strings.LastIndex(forVal.UserEmail, "@")+1:],
				"admin":  false,
			}
			break
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		attr, err = s.getAttributesFor(ctx, user.ID, proj.OrganizationID, proj.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	case *adminv1.GetIFrameRequest_Attributes:
		attr = forVal.Attributes.AsMap()
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
	}

	ttlDuration := 24 * time.Hour
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
				// TODO: Remove ReadProfiling and ReadRepo (may require frontend changes)
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadProfiling,
				runtimeauth.ReadRepo,
			},
		},
		Attributes: attr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(prodDepl.ID)

	if req.Kind == "" {
		req.Kind = "MetricsView"
	}

	iFrameURL, err := urlutil.WithQuery(urlutil.MustJoinURL(s.opts.FrontendURL, "/-/embed"), map[string]string{
		"runtime_host": prodDepl.RuntimeHost,
		"instance_id":  prodDepl.RuntimeInstanceID,
		"access_token": jwt,
		"kind":         req.Kind,
		"resource":     req.Resource,
		"state":        "",
		"theme":        req.Query["theme"],
	})
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
