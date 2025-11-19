package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
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
		InstanceId: depl.RuntimeInstanceID,
		Resources:  names,
		All:        len(names) == 0, // Backwards compatibility
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRefreshSourcesResponse{}, nil
}

// ListDeployments returns a list of deployments for a given project.
func (s *Server) ListDeployments(ctx context.Context, req *adminv1.ListDeploymentsRequest) (*adminv1.ListDeploymentsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization_name", req.Org),
		attribute.String("args.project_name", req.Project),
		attribute.String("args.environment", req.Environment),
		attribute.String("args.user_id", req.UserId),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	depls, err := s.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Filter deployments based on permissions and specified environment and user ID.
	var newDepls []*database.Deployment
	for _, d := range depls {
		if d.Environment == "prod" && !permissions.ReadProd {
			continue
		}
		if d.Environment == "dev" && !permissions.ReadDev {
			continue
		}
		if req.Environment != "" && req.Environment != d.Environment {
			continue
		}
		if req.UserId != "" && d.OwnerUserID != nil && req.UserId != *d.OwnerUserID {
			continue
		}
		newDepls = append(newDepls, d)
	}

	dtos := make([]*adminv1.Deployment, len(newDepls))
	for i, d := range newDepls {
		dtos[i] = deploymentToDTO(d)
	}

	return &adminv1.ListDeploymentsResponse{
		Deployments: dtos,
	}, nil
}

// GetDeployment returns runtime info and JWT on behalf of a specific user, or alternatively for a raw set of JWT attributes
func (s *Server) GetDeployment(ctx context.Context, req *adminv1.GetDeploymentRequest) (*adminv1.GetDeploymentResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.deployment_id", req.DeploymentId),
		attribute.String("args.access_token_ttl_seconds", strconv.FormatUint(uint64(req.AccessTokenTtlSeconds), 10)),
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
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if depl.Environment == "dev" {
		if !permissions.ReadDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to read dev deployment")
		}
	} else {
		if !permissions.ReadProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to read prod deployment")
		}
	}

	var attr map[string]any
	if req.For == nil {
		if claims.OwnerType() == auth.OwnerTypeUser {
			attr, err = s.jwtAttributesForUser(ctx, claims.OwnerID(), proj.OrganizationID, permissions)
			if err != nil {
				return nil, err
			}
		} else if claims.OwnerType() == auth.OwnerTypeService {
			attr = map[string]any{"admin": true}
		}
	} else {
		if depl.Environment == "prod" && !permissions.ManageProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage prod deployment")
		}

		if depl.Environment == "dev" && !permissions.ManageDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage dev deployment")
		}

		switch forVal := req.For.(type) {
		case *adminv1.GetDeploymentRequest_UserId:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, forVal.UserId, "")
			if err != nil {
				return nil, err
			}
		case *adminv1.GetDeploymentRequest_UserEmail:
			attr, err = s.getAttributesForUser(ctx, proj.OrganizationID, proj.ID, "", forVal.UserEmail)
			if err != nil {
				return nil, err
			}
		case *adminv1.GetDeploymentRequest_Attributes:
			attr = forVal.Attributes.AsMap()
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid 'for' type")
		}
	}

	ttlDuration := runtimeAccessTokenEmbedTTL
	if req.AccessTokenTtlSeconds > 0 {
		ttlDuration = time.Duration(req.AccessTokenTtlSeconds) * time.Second
	}

	instancePermissions := []runtime.Permission{
		runtime.ReadObjects,
		runtime.ReadMetrics,
		runtime.ReadAPI,
		runtime.UseAI,
	}
	if depl.Environment == "dev" {
		instancePermissions = append(instancePermissions,
			runtime.ReadOLAP,
			runtime.ReadProfiling,
			runtime.ReadRepo,
			runtime.ReadResolvers,
		)
		if permissions.ManageDev {
			instancePermissions = append(instancePermissions,
				runtime.EditRepo,
				runtime.EditTrigger,
			)
		}
	} else if permissions.ManageProd {
		instancePermissions = append(instancePermissions,
			runtime.ReadResolvers,
			runtime.EditTrigger,
		)
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: depl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         ttlDuration,
		InstancePermissions: map[string][]runtime.Permission{
			depl.RuntimeInstanceID: instancePermissions,
		},
		Attributes: attr,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(depl.ID)

	return &adminv1.GetDeploymentResponse{
		RuntimeHost: depl.RuntimeHost,
		InstanceId:  depl.RuntimeInstanceID,
		AccessToken: jwt,
		TtlSeconds:  uint32(ttlDuration.Seconds()),
	}, nil
}

// CreateDeployment creates a new deployment for a project.
func (s *Server) CreateDeployment(ctx context.Context, req *adminv1.CreateDeploymentRequest) (*adminv1.CreateDeploymentResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization_name", req.Org),
		attribute.String("args.project_name", req.Project),
		attribute.String("args.environment", req.Environment),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if req.Environment == "dev" {
		if !permissions.ManageDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage dev deployment")
		}
	} else {
		if !permissions.ManageProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage prod deployment")
		}
	}

	// We only allow one prod deployment.
	if req.Environment == "prod" && proj.ProdDeploymentID != nil {
		return nil, status.Error(codes.InvalidArgument, "project already has a prod deployment, cannot create a new prod deployment")
	}

	// Determine branch and slots based on environment.
	var branch string
	var slots int
	switch req.Environment {
	case "prod":
		branch = ""
		slots = proj.ProdSlots
	case "dev":
		// Generate a random branch name for dev deployments
		b := make([]byte, 8)
		_, err := rand.Read(b)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		branch = fmt.Sprintf("rill/%s", hex.EncodeToString(b))
		slots = proj.DevSlots
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid environment specified, must be 'prod' or 'dev'")
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check projects quota
	usage, err := s.admin.DB.CountProjectsQuotaUsage(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	if org.QuotaSlotsPerDeployment >= 0 && slots > org.QuotaSlotsPerDeployment {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can't provision more than %d slots per deployment", org.QuotaSlotsPerDeployment)
	}
	if org.QuotaSlotsTotal >= 0 && usage.Slots+slots > org.QuotaSlotsTotal {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d total slots", org.Name, org.QuotaSlotsTotal)
	}
	if org.QuotaDeployments >= 0 && usage.Deployments >= org.QuotaDeployments {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d deployments", org.Name, org.QuotaDeployments)
	}

	// If the request is for a dev deployment and the owner is a user, we set the ownerUserID
	var ownerUserID *string
	if req.Environment == "dev" && claims.OwnerType() == auth.OwnerTypeUser {
		id := claims.OwnerID()
		ownerUserID = &id
	}

	depl, err := s.admin.CreateDeployment(ctx, &admin.CreateDeploymentOptions{
		ProjectID:   proj.ID,
		OwnerUserID: ownerUserID,
		Environment: req.Environment,
		Branch:      branch,
	})
	if err != nil {
		return nil, err
	}

	if depl.Environment == "prod" {
		// If this is a prod deployment, we update the prod deployment on project
		_, err = s.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
			Name:                 proj.Name,
			Description:          proj.Description,
			Public:               proj.Public,
			DirectoryName:        proj.DirectoryName,
			Provisioner:          proj.Provisioner,
			ArchiveAssetID:       proj.ArchiveAssetID,
			GitRemote:            proj.GitRemote,
			GithubInstallationID: proj.GithubInstallationID,
			GithubRepoID:         proj.GithubRepoID,
			ManagedGitRepoID:     proj.ManagedGitRepoID,
			ProdVersion:          proj.ProdVersion,
			ProdBranch:           proj.ProdBranch,
			Subpath:              proj.Subpath,
			ProdDeploymentID:     &depl.ID,
			ProdSlots:            proj.ProdSlots,
			ProdTTLSeconds:       proj.ProdTTLSeconds,
			DevSlots:             proj.DevSlots,
			DevTTLSeconds:        proj.DevTTLSeconds,
			Annotations:          proj.Annotations,
		})
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.CreateDeploymentResponse{
		Deployment: deploymentToDTO(depl),
	}, nil
}

// StartDeployment starts a deployment by ID.
func (s *Server) StartDeployment(ctx context.Context, req *adminv1.StartDeploymentRequest) (*adminv1.StartDeploymentResponse, error) {
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
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if depl.Environment == "dev" {
		if !permissions.ManageDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage dev deployment")
		}
	} else {
		if !permissions.ManageProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage prod deployment")
		}
	}

	depl, err = s.admin.StartDeployment(ctx, depl)
	if err != nil {
		return nil, err
	}

	s.admin.Used.Deployment(depl.ID)

	return &adminv1.StartDeploymentResponse{
		Deployment: deploymentToDTO(depl),
	}, nil
}

// StopDeployment stops a deployment by ID.
func (s *Server) StopDeployment(ctx context.Context, req *adminv1.StopDeploymentRequest) (*adminv1.StopDeploymentResponse, error) {
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
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if depl.Environment == "dev" {
		if !permissions.ManageDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage dev deployment")
		}
	} else {
		if !permissions.ManageProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage prod deployment")
		}
	}

	err = s.admin.StopDeployment(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.StopDeploymentResponse{
		DeploymentId: depl.ID,
	}, nil
}

// DeleteDeployment deletes a deployment by ID.
func (s *Server) DeleteDeployment(ctx context.Context, req *adminv1.DeleteDeploymentRequest) (*adminv1.DeleteDeploymentResponse, error) {
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
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if depl.Environment == "dev" {
		if !permissions.ManageDev {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage dev deployment")
		}
	} else {
		if !permissions.ManageProd {
			return nil, status.Error(codes.PermissionDenied, "does not have permission to manage prod deployment")
		}
	}

	err = s.admin.TeardownDeployment(ctx, depl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.DeleteDeploymentResponse{DeploymentId: depl.ID}, nil
}

// GetDeploymentCredentials returns runtime info and JWT on behalf of a specific user, or alternatively for a raw set of JWT attributes
func (s *Server) GetDeploymentCredentials(ctx context.Context, req *adminv1.GetDeploymentCredentialsRequest) (*adminv1.GetDeploymentCredentialsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.branch", req.Branch),
		attribute.String("args.ttl_seconds", strconv.FormatUint(uint64(req.TtlSeconds), 10)),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
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

	if !permissions.ManageProd {
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
		InstancePermissions: map[string][]runtime.Permission{
			prodDepl.RuntimeInstanceID: {
				runtime.ReadObjects,
				runtime.ReadMetrics,
				runtime.ReadAPI,
				runtime.UseAI,
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
		attribute.String("args.organization", req.Org),
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

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
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
	s.admin.Used.Deployment(prodDepl.ID)

	if req.Branch != "" && req.Branch != prodDepl.Branch {
		return nil, status.Error(codes.InvalidArgument, "project does not have a deployment for given branch")
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)

	if !permissions.ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	// Get user attributes to pass in the JWT
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

	// Add an `embed` attribute for use in security policies or feature flags (as `{{.user.embed}}`).
	if _, ok := attr["embed"]; !ok {
		if attr == nil {
			attr = make(map[string]any)
		}
		attr["embed"] = true
	}

	// Backwards compatibility for req.Type and req.Kind
	if req.Kind != "" { // nolint:staticcheck // For backwards compatibility
		req.Type = req.Kind // nolint:staticcheck // For backwards compatibility
	}
	if req.Type == "" {
		// Default to an explore if no type is explicitly provided
		req.Type = runtime.ResourceKindExplore
	}
	req.Type = runtime.ResourceKindFromShorthand(req.Type)

	// If navigation is disabled and a specific resource is requested, limit access to only that resource.
	var rules []*runtimev1.SecurityRule
	if !req.Navigation && req.Resource != "" {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_TransitiveAccess{
				TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
					Resource: &runtimev1.ResourceName{
						Kind: req.Type,
						Name: req.Resource,
					},
				},
			},
		})
	}

	// Determine TTL for the access token
	ttlDuration := runtimeAccessTokenEmbedTTL
	if req.TtlSeconds > 0 {
		ttlDuration = time.Duration(req.TtlSeconds) * time.Second
	}

	// Generate JWT
	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: prodDepl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         ttlDuration,
		InstancePermissions: map[string][]runtime.Permission{
			prodDepl.RuntimeInstanceID: {
				runtime.ReadObjects,
				runtime.ReadMetrics,
				runtime.ReadAPI,
				runtime.UseAI,
			},
		},
		Attributes:    attr,
		SecurityRules: rules,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	// Build the iframe URL search params
	iframeQuery := map[string]string{
		"runtime_host": prodDepl.RuntimeHost,
		"instance_id":  prodDepl.RuntimeInstanceID,
		"access_token": jwt,
	}

	iframeQuery["type"] = req.Type
	iframeQuery["kind"] = req.Type // For backwards compatibility

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
