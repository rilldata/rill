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

// issueRuntimeTokenOptions configures a call to issueRuntimeToken.
type issueRuntimeTokenOptions struct {
	// project the deployment belongs to.
	project *database.Project
	// deployment the token will be used for.
	deployment *database.Deployment
	// projectPermissions of the claims owner.
	// Passed separately from the claims to accommodate overrides for public projects and forced superuser access.
	projectPermissions *adminv1.ProjectPermissions
	// forOwner issues the token for the current claims owner (user/service/anon/etc).
	forOwner bool
	// forUserID issues the token for a specific Rill user. Mutually exclusive with the other for* fields.
	forUserID string
	// forUserEmail issues the token for a user by email.
	// The email does not have to correspond to an existing user; if it doesn't, synthetic attributes are generated based on the email.
	// Mutually exclusive with the other for* fields.
	forUserEmail string
	// forUserAttributes issues the token with explicit user attributes (no extra resolution).
	// A non-nil empty map counts as set and selects this mode.
	// Mutually exclusive with the other for* fields.
	forUserAttributes map[string]any
	// externalUserID is an optional external user ID to be used when the token is issued for a non-Rill end user (usually in an embedded context).
	// It will be hashed and used as the JWT's subject.
	// It cannot be combined with forOwner or forUserID.
	// It may be set on its own, i.e. you do not have to set any of the for* fields.
	externalUserID string
	// grantManageAll grants runtime.ManageInstances permission.
	// Only available for tokens with forOwner where the owner is a superuser.
	grantManageAll bool
	// extraAttributes will be merged into the resolved JWT attributes, but will not override attributes from other sources.
	extraAttributes map[string]any
	// overrideResources optionally overrides and replaces other resource-restriction rules.
	overrideResources []database.ResourceName
	// ttl is how long the token is valid for.
	ttl time.Duration
}

// issueRuntimeToken issues a runtime JWT for a deployment based on the provided options and the currently authenticated user.
// It should only be used from server handlers as it requires the context to carry auth claims.
func (s *Server) issueRuntimeToken(ctx context.Context, opts *issueRuntimeTokenOptions) (string, error) {
	// Get claims
	claims := auth.GetClaims(ctx)
	if claims == nil {
		return "", status.Error(codes.Unauthenticated, "cannot issue runtime token without claims")
	}

	// Validate that at most one "for" mode is set.
	// Note: it is valid to set none of them, which means no base subject/attributes/rules will be resolved.
	// Note: forManagement is treated differently as it is a modifier on forOwner, which is checked separately below.
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
		return "", status.Error(codes.Internal, "at most one of forOwner/forUserID/forUserEmail/forUserAttributes may be set")
	}

	// Resolve principal: subject, attributes, and resource-restriction rules.
	var subject string
	var attr map[string]any
	var rules []*runtimev1.SecurityRule
	switch {
	case opts.forOwner:
		if opts.externalUserID != "" {
			return "", status.Error(codes.Internal, "externalUserID cannot be specified together with forOwner")
		}
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
	case opts.forUserID != "":
		if opts.externalUserID != "" {
			return "", status.Error(codes.Internal, "externalUserID cannot be specified together with forUserID")
		}
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
	default:
		// No principal: no subject, no attributes, no rules.
	}

	// Apply externalUserID after principal resolution so it overrides whatever subject was set.
	if opts.externalUserID != "" {
		subject = subjectForExternalUser(opts.externalUserID, opts.project.ID)
	}

	// overrideResources replaces any principal-derived resource rules with a fresh allow-list.
	if len(opts.overrideResources) > 0 {
		rules = securityRulesFromResources(true, opts.overrideResources)
	}

	// Merge extraAttributes (earlier keys win)
	if len(opts.extraAttributes) > 0 {
		if attr == nil {
			attr = make(map[string]any, len(opts.extraAttributes))
		}
		for k, v := range opts.extraAttributes {
			if _, exists := attr[k]; !exists {
				attr[k] = v
			}
		}
	}

	// Check if allowed to manage the deployment's environment.
	// NOTE: Only applicable for tokens issued for the claims owner (not possible to delegate to other end users).
	var manageDepl bool
	if opts.forOwner {
		if opts.deployment.Environment == "prod" {
			manageDepl = opts.projectPermissions.ManageProd
		} else {
			manageDepl = opts.projectPermissions.ManageDev
		}
	}

	// Derive instance permissions from deployment config and project permissions.
	instancePermissions := []runtime.Permission{
		runtime.ReadAPI,
		runtime.ReadMetrics,
		runtime.ReadObjects,
		runtime.UseAI,
	}
	if manageDepl {
		instancePermissions = append(
			instancePermissions,
			runtime.ReadInstance,
			runtime.ReadResolvers,
			runtime.EditTrigger,
		)
		if opts.deployment.Editable {
			instancePermissions = append(
				instancePermissions,
				runtime.ReadOLAP,
				runtime.ReadProfiling,
				runtime.ReadRepo,
				runtime.EditRepo,
			)
		}
	}

	// Derive system permissions; only used for grantManageAll for now.
	var systemPermissions []runtime.Permission
	if opts.grantManageAll {
		// Must be for an owner who is a superuser
		if !opts.forOwner {
			return "", status.Error(codes.Internal, "grantManageAll requires forOwner")
		}
		if !claims.Superuser(ctx) {
			return "", status.Error(codes.PermissionDenied, "only superusers can issue management tokens")
		}

		// NOTE: ManageInstances is currently used by the runtime to allow skipping access checks.
		systemPermissions = append(systemPermissions, runtime.ManageInstances)
	}

	// Issue JWT.
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
