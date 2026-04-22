package server

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// issueRuntimeTokenOptions configures a call to (*Server).issueRuntimeToken.
type issueRuntimeTokenOptions struct {
	// project the deployment belongs to.
	project *database.Project
	// deployment the token will be used for.
	deployment *database.Deployment
	// projectPermissions of the claims owner.
	// Passed separately from the claims to accommodate overrides for public projects and forced superuser access.
	projectPermissions *adminv1.ProjectPermissions
	// forOwner issues the token for the current claims owner (user / service / anon / magic token).
	forOwner bool
	// forManagement grants runtime.ManageInstances. May only be combined with forOwner.
	// The helper verifies the caller is a superuser and otherwise returns PermissionDenied.
	forManagement bool
	// forUserID issues the token for a specific Rill user. Mutually exclusive with the other for* fields.
	forUserID string
	// forUserEmail issues the token for a user by email (synthesising non-admin attributes if unknown).
	// Mutually exclusive with the other for* fields.
	forUserEmail string
	// forUserAttributes issues the token with explicit attributes (no principal resolution).
	// Mutually exclusive with the other for* fields.
	forUserAttributes map[string]any
	// externalUserID, if non-empty, sets the JWT subject to a stable hash of the id scoped to the project.
	// May be set on its own — in that case no principal is resolved and the token carries only the
	// external subject (plus any extraAttributes). Cannot be combined with forUserID, whose user
	// ID would itself be the subject.
	externalUserID string
	// extraAttributes is merged into the JWT attributes and takes precedence over principal-derived keys.
	extraAttributes map[string]any
	// overrideResources, when non-nil, replaces principal-derived resource-restriction rules with
	// TransitiveAccess rules for exactly these resources. Incompatible with magic-token principals.
	overrideResources []*runtimev1.ResourceName
	// ttl is how long the token is valid for.
	ttl time.Duration
}

// issueRuntimeToken issues a runtime JWT for a deployment based on opts.
// It centralizes subject/attribute/rule resolution and instance-permission derivation
// that was previously duplicated across GetProject, GetDeployment, GetDeploymentCredentials,
// GetIFrame, and the runtime proxy.
func (s *Server) issueRuntimeToken(ctx context.Context, opts *issueRuntimeTokenOptions) (string, error) {
	// Validate that exactly one "for" mode is set.
	forCount := 0
	if opts.forOwner {
		forCount++
	}
	if opts.forUserID != "" {
		forCount++
	}
	if opts.forUserEmail != "" {
		forCount++
	}
	if opts.forUserAttributes != nil {
		forCount++
	}
	if forCount > 1 {
		return "", status.Error(codes.Internal, "issueRuntimeToken: at most one of forOwner/forUserID/forUserEmail/forUserAttributes may be set")
	}
	if opts.forManagement && !opts.forOwner {
		return "", status.Error(codes.Internal, "issueRuntimeToken: forManagement requires forOwner")
	}
	if opts.externalUserID != "" && opts.forUserID != "" {
		return "", status.Error(codes.InvalidArgument, "external_user_id cannot be specified together with user_id")
	}

	claims := auth.GetClaims(ctx)

	// Enforce forManagement: only superusers may mint management tokens.
	if opts.forManagement && !claims.Superuser(ctx) {
		return "", status.Error(codes.PermissionDenied, "only superusers can issue management tokens")
	}

	// Resolve principal: subject, attributes, and resource-restriction rules.
	var subject string
	var attr map[string]any
	var rules []*runtimev1.SecurityRule
	switch {
	case opts.forUserID != "":
		subject = opts.forUserID
		a, restrict, resources, err := s.getAttributesAndResourceRestrictionsForUser(ctx, opts.project.OrganizationID, opts.project.ID, opts.forUserID, "")
		if err != nil {
			return "", err
		}
		attr = a
		rules = append(rules, securityRulesFromResources(restrict, resources)...)
	case opts.forUserEmail != "":
		a, restrict, resources, err := s.getAttributesAndResourceRestrictionsForUser(ctx, opts.project.OrganizationID, opts.project.ID, "", opts.forUserEmail)
		if err != nil {
			return "", err
		}
		attr = a
		rules = append(rules, securityRulesFromResources(restrict, resources)...)
	case opts.forUserAttributes != nil:
		attr = opts.forUserAttributes
	case opts.forOwner:
		switch claims.OwnerType() {
		case auth.OwnerTypeUser:
			subject = claims.OwnerID()
			a, err := s.jwtAttributesForUser(ctx, claims.OwnerID(), opts.project.OrganizationID, opts.projectPermissions)
			if err != nil {
				return "", err
			}
			attr = a
			restrict, resources, err := s.getResourceRestrictionsForUser(ctx, opts.project.ID, claims.OwnerID())
			if err != nil {
				return "", err
			}
			rules = append(rules, securityRulesFromResources(restrict, resources)...)
		case auth.OwnerTypeService:
			subject = claims.OwnerID()
			a, err := s.jwtAttributesForService(ctx, claims.OwnerID(), opts.projectPermissions)
			if err != nil {
				return "", err
			}
			attr = a
		case auth.OwnerTypeMagicAuthToken:
			if opts.overrideResources != nil {
				return "", status.Error(codes.PermissionDenied, "resource overrides are not supported for magic-token credentials")
			}
			subject = claims.OwnerID()
			mdl, ok := claims.AuthTokenModel().(*database.MagicAuthToken)
			if !ok {
				return "", status.Errorf(codes.Internal, "unexpected type %T for magic auth token model", claims.AuthTokenModel())
			}
			attr = mdl.Attributes
			magicRules, err := securityRulesFromMagicAuthToken(mdl)
			if err != nil {
				return "", err
			}
			rules = append(rules, magicRules...)
		case auth.OwnerTypeAnon:
			// Anonymous principal: no subject, no attributes, no rules.
		default:
			return "", status.Errorf(codes.InvalidArgument, "unsupported owner type %q", claims.OwnerType())
		}
	}

	// Apply externalUserID after principal resolution so it overrides whatever subject was set.
	if opts.externalUserID != "" {
		subject = subjectForExternalUser(opts.externalUserID, opts.project.ID)
	}

	// overrideResources replaces any principal-derived resource rules with a fresh allow-list.
	// Magic-token principals have already been rejected above.
	if opts.overrideResources != nil {
		rules = rules[:0]
		for _, r := range opts.overrideResources {
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: r,
					},
				},
			})
		}
	}

	// Merge extraAttributes; later keys win so callers can force specific values (e.g., "embed": true).
	if len(opts.extraAttributes) > 0 {
		if attr == nil {
			attr = make(map[string]any, len(opts.extraAttributes))
		}
		for k, v := range opts.extraAttributes {
			attr[k] = v
		}
	}

	// Derive instance permissions uniformly from deployment environment + project permissions.
	instancePermissions := []runtime.Permission{
		runtime.ReadObjects,
		runtime.ReadMetrics,
		runtime.ReadAPI,
		runtime.UseAI,
	}
	if opts.deployment.Environment == "dev" {
		instancePermissions = append(instancePermissions,
			runtime.ReadOLAP,
			runtime.ReadProfiling,
			runtime.ReadRepo,
			runtime.ReadResolvers,
		)
		if opts.projectPermissions.ManageDev {
			instancePermissions = append(instancePermissions, runtime.EditRepo, runtime.EditTrigger)
		}
	} else {
		if opts.projectPermissions.ManageProd || opts.projectPermissions.ManageProject {
			instancePermissions = append(instancePermissions, runtime.ReadResolvers, runtime.EditTrigger)
		}
		if opts.projectPermissions.ManageProject {
			instancePermissions = append(instancePermissions, runtime.ReadInstance)
		}
	}

	var systemPermissions []runtime.Permission
	if opts.forManagement {
		// NOTE: ManageInstances is currently used by the runtime to skip access checks.
		systemPermissions = append(systemPermissions, runtime.ManageInstances)
	}

	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL:       opts.deployment.RuntimeAudience,
		Subject:           subject,
		TTL:               opts.ttl,
		SystemPermissions: systemPermissions,
		InstancePermissions: map[string][]runtime.Permission{
			opts.deployment.RuntimeInstanceID: instancePermissions,
		},
		Attributes:    attr,
		SecurityRules: rules,
	})
	if err != nil {
		return "", status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	return jwt, nil
}

// securityRulesFromMagicAuthToken builds the security rules encoded by a magic auth token:
// a resource allow-list (or blanket deny), metrics-view row filters, and a field allow-list.
func securityRulesFromMagicAuthToken(mdl *database.MagicAuthToken) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule

	if len(mdl.Resources) == 0 {
		// No resources means deny all access.
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{Allow: false},
			},
		})
	} else {
		for _, r := range mdl.Resources {
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: r.Type,
							Name: r.Name,
						},
					},
				},
			})
		}
	}

	for mv, filter := range mdl.MetricsViewFilterJSONs {
		if mv == "" {
			return nil, status.Errorf(codes.Internal, "empty metrics view name in metrics view filter")
		}
		expr := &runtimev1.Expression{}
		if err := protojson.Unmarshal([]byte(filter), expr); err != nil {
			return nil, status.Errorf(codes.Internal, "could not unmarshal metrics view %q filter: %s", mv, err.Error())
		}
		if mv == "*" {
			// Backwards compatibility: "*" applies to all metrics views.
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_RowFilter{
					RowFilter: &runtimev1.SecurityRuleRowFilter{
						Expression: expr,
					},
				},
			})
			continue
		}
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_RowFilter{
				RowFilter: &runtimev1.SecurityRuleRowFilter{
					ConditionResources: []*runtimev1.ResourceName{{
						Kind: runtime.ResourceKindMetricsView,
						Name: mv,
					}},
					Expression: expr,
				},
			},
		})
	}

	if len(mdl.Fields) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					Fields:    mdl.Fields,
					Allow:     true,
					Exclusive: true,
				},
			},
		})
	}

	return rules, nil
}
