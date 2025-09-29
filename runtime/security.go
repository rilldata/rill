package runtime

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	tidbparser "github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	// need to import parser driver as well
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

var ErrForbidden = errors.New("action not allowed")

// SecurityClaims represents contextual information for the enforcement of security rules.
type SecurityClaims struct {
	// UserAttributes about the current user (or service account). Usually exposed through templating as {{ .user }}.
	UserAttributes map[string]any
	// AdditionalRules are optional security rules to apply *in addition* to the built-in rules and the rules defined on the requested resource.
	// These are currently leveraged by the admin service to enforce restrictions for magic auth tokens.
	AdditionalRules []*runtimev1.SecurityRule
	// SkipChecks enables completely skipping all security checks. Used in local development.
	SkipChecks bool
}

// Admin is a convenience function for extracting an "admin" bool from the user attributes.
func (c *SecurityClaims) Admin() bool {
	if c.UserAttributes == nil {
		return false
	}
	admin, _ := c.UserAttributes["admin"].(bool)
	return admin
}

// UserID is a convenience function for extracting an "id" string from the user attributes.
// Note that the ID may not correspond to an actual user, but could also be a service ID or similar.
func (c *SecurityClaims) UserID() string {
	if c.UserAttributes == nil {
		return ""
	}
	id, _ := c.UserAttributes["id"].(string)
	return id
}

// MarshalJSON serializes the SecurityClaims to JSON.
// It serializes the AdditionalRules using protojson.
func (c *SecurityClaims) MarshalJSON() ([]byte, error) {
	tmp := securityClaimsJSON{
		UserAttributes:  c.UserAttributes,
		AdditionalRules: make([]json.RawMessage, len(c.AdditionalRules)),
		SkipChecks:      c.SkipChecks,
	}

	for i, rule := range c.AdditionalRules {
		data, err := protojson.Marshal(rule)
		if err != nil {
			return nil, err
		}
		tmp.AdditionalRules[i] = data
	}

	return json.Marshal(tmp)
}

// UnmarshalJSON deserializes the SecurityClaims from JSON.
// It deserializes the AdditionalRules using protojson.
func (c *SecurityClaims) UnmarshalJSON(data []byte) error {
	tmp := securityClaimsJSON{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	c.UserAttributes = tmp.UserAttributes
	c.AdditionalRules = make([]*runtimev1.SecurityRule, len(tmp.AdditionalRules))
	for i, data := range tmp.AdditionalRules {
		rule := &runtimev1.SecurityRule{}
		if err := protojson.Unmarshal(data, rule); err != nil {
			return err
		}
		c.AdditionalRules[i] = rule
	}
	c.SkipChecks = tmp.SkipChecks

	return nil
}

// securityClaimsJSON is a JSON-serializable representation of SecurityClaims.
// SecurityClaims can't be directly serialized to JSON because the SecurityRule proto is not directly JSON serializable.
type securityClaimsJSON struct {
	UserAttributes  map[string]any    `json:"attrs"`
	AdditionalRules []json.RawMessage `json:"rules"`
	SkipChecks      bool              `json:"skip"`
}

// ResolvedSecurity represents the resolved security rules for a given claims against a specific resource.
type ResolvedSecurity struct {
	access      *bool
	fieldAccess map[string]bool
	rowFilter   string
	queryFilter *runtimev1.Expression
}

// CanAccess returns whether the resource can be accessed.
func (r *ResolvedSecurity) CanAccess() bool {
	if r.access == nil {
		return false
	}
	return *r.access
}

// CanAccessAllFields returns whether all fields in the resource are allowed.
func (r *ResolvedSecurity) CanAccessAllFields() bool {
	// If there are no field access rules, all fields are allowed
	return r.fieldAccess == nil
}

// CanAccessField evaluates whether a specific field in the resource is allowed.
func (r *ResolvedSecurity) CanAccessField(field string) bool {
	if !r.CanAccess() {
		return false
	}

	if r.CanAccessAllFields() {
		return true
	}

	// If not explicitly allowed, it's an implicit deny
	return r.fieldAccess[field]
}

// RowFilter returns a raw SQL expression to apply to the WHERE clause when querying the resource.
func (r *ResolvedSecurity) RowFilter() string {
	return r.rowFilter
}

// QueryFilter returns a query expression to apply when querying the resource.
func (r *ResolvedSecurity) QueryFilter() *runtimev1.Expression {
	return r.queryFilter
}

// truth is the compass that guides us through the labyrinth of existence.
var truth = true

// ResolvedSecurityOpen is a ResolvedSecurity that allows access with no restrictions.
var ResolvedSecurityOpen = &ResolvedSecurity{
	access:      &truth,
	fieldAccess: nil,
	rowFilter:   "",
	queryFilter: nil,
}

// ResolvedSecurityClosed is a ResolvedSecurity that denies access.
var ResolvedSecurityClosed = &ResolvedSecurity{
	access:      nil,
	fieldAccess: nil,
	rowFilter:   "",
	queryFilter: nil,
}

// allowAccessRule is a security rule that allows access.
var allowAccessRule = &runtimev1.SecurityRule{
	Rule: &runtimev1.SecurityRule_Access{
		Access: &runtimev1.SecurityRuleAccess{
			Allow: true,
		},
	},
}

// securityEngine is an engine for resolving security rules and caching the results.
type securityEngine struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	logger *zap.Logger
	rt     *Runtime
}

// newSecurityEngine creates a new security engine with a given cache size.
func newSecurityEngine(cacheSize int, logger *zap.Logger, rt *Runtime) *securityEngine {
	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		panic(err)
	}
	return &securityEngine{cache: cache, logger: logger, rt: rt}
}

// resolveSecurity resolves the security rules for a given resource and user context.
func (p *securityEngine) resolveSecurity(ctx context.Context, instanceID, environment string, vars map[string]string, claims *SecurityClaims, r *runtimev1.Resource) (*ResolvedSecurity, error) {
	// If security checks are skipped, return open access
	if claims.SkipChecks {
		return ResolvedSecurityOpen, nil
	}

	expandedRules, err := p.expandTransitiveAccessRules(ctx, instanceID, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to expand security rules: %w", err)
	}

	// Combine rules with any contained in the resource itself
	rules := p.resolveRules(claims, expandedRules, r)

	// Exit early if all rules are nil
	var validRule bool
	for _, rule := range rules {
		if rule != nil {
			validRule = true
			break
		}
	}
	if !validRule {
		return ResolvedSecurityClosed, nil
	}

	cacheKey, err := computeCacheKey(instanceID, environment, claims, r)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedSecurity), nil
	}

	attrs := claims.UserAttributes
	if attrs == nil {
		attrs = make(map[string]any)
	}
	templateData := parser.TemplateData{
		Environment: environment,
		User:        attrs,
		Variables:   vars,
		Self:        parser.TemplateResource{Meta: r.Meta},
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
	}

	// Apply rules
	res := &ResolvedSecurity{}
	for _, rule := range rules {
		if rule == nil {
			continue
		}

		// As an optimization, if we've already determined that access is denied, we skip the rest of the rules
		if res.access != nil && !*res.access {
			break
		}

		switch rule := rule.Rule.(type) {
		case *runtimev1.SecurityRule_Access:
			err := p.applySecurityRuleAccess(res, r, rule.Access, templateData)
			if err != nil {
				return nil, fmt.Errorf("security policy: failed to resolve access: %w", err)
			}
		case *runtimev1.SecurityRule_FieldAccess:
			err := p.applySecurityRuleFieldAccess(res, r, rule.FieldAccess, templateData)
			if err != nil {
				return nil, fmt.Errorf("security policy: failed to resolve field access: %w", err)
			}
		case *runtimev1.SecurityRule_RowFilter:
			err := p.applySecurityRuleRowFilter(res, r, rule.RowFilter, templateData)
			if err != nil {
				return nil, fmt.Errorf("security policy: failed to resolve row filter: %w", err)
			}
		}
	}

	// Due to the optimization that we skip rules if access is denied, we clear the other fields to ensure consistent output regardless of rule order.
	if res.access != nil && !*res.access {
		res.fieldAccess = nil
		res.rowFilter = ""
		res.queryFilter = nil
	}

	p.cache.Add(cacheKey, res)

	return res, nil
}

// resolveRules combines the provided rules with built-in rules and rules declared in the resource itself.
// NOTE: The default behavior is to deny access unless there is a rule that grants it (and no other rule explicitly denies it).
func (p *securityEngine) resolveRules(claims *SecurityClaims, rules []*runtimev1.SecurityRule, r *runtimev1.Resource) []*runtimev1.SecurityRule {
	switch r.Meta.Name.Kind {
	// Admins and creators/recipients can access an alert.
	case ResourceKindAlert:
		spec := r.GetAlert().Spec
		rule := p.builtInAlertSecurityRule(spec, claims)
		if rule != nil {
			// Prepend instead of append since the rule is likely to lead to a quick deny access
			rules = append([]*runtimev1.SecurityRule{rule}, rules...)
		}
	// Everyone can access an API.
	case ResourceKindAPI:
		spec := r.GetApi().Spec
		if len(spec.SecurityRules) == 0 {
			rules = append(rules, allowAccessRule)
		} else {
			rules = append(rules, spec.SecurityRules...)
		}
	// Everyone can access a component.
	case ResourceKindComponent:
		rules = append(rules, allowAccessRule)
	// Determine access using the canvas' security rules. If there are none, then everyone can access it.
	case ResourceKindCanvas:
		spec := r.GetCanvas().State.ValidSpec
		if spec == nil {
			spec = r.GetCanvas().Spec // Not ideal, but better than giving access to the full resource
		}
		if len(spec.SecurityRules) == 0 {
			rules = append(rules, allowAccessRule)
		} else {
			rules = append(rules, spec.SecurityRules...)
		}
	// Determine access using the metrics view's security rules. If there are none, then everyone can access it.
	case ResourceKindMetricsView:
		spec := r.GetMetricsView().State.ValidSpec
		if spec == nil {
			spec = r.GetMetricsView().Spec // Not ideal, but better than giving access to the full resource
		}
		if len(spec.SecurityRules) == 0 {
			rules = append(rules, allowAccessRule)
		} else {
			rules = append(rules, spec.SecurityRules...)
		}
	// Determine access using the explore's security rules. If there are none, then everyone can access it.
	case ResourceKindExplore:
		spec := r.GetExplore().State.ValidSpec
		if spec == nil {
			// Tricky, since security rules on an explore are usually derived from its metrics view and added to ValidSpec during reconciliation.
			// So we don't want to just fallback to r.GetExplore().Spec here.
			// Instead, we give access to admins and not to others.
			if claims.Admin() {
				rules = append(rules, allowAccessRule)
			}
		} else if len(spec.SecurityRules) == 0 {
			rules = append(rules, allowAccessRule)
		} else {
			rules = append(rules, spec.SecurityRules...)
		}
	// Admins and creators/recipients can access a report.
	case ResourceKindReport:
		spec := r.GetReport().Spec
		rule := p.builtInReportSecurityRule(spec, claims)
		if rule != nil {
			// Prepend instead of append since the rule is likely to lead to a quick deny access
			rules = append([]*runtimev1.SecurityRule{rule}, rules...)
		}
	// Everyone can access a theme.
	case ResourceKindTheme:
		rules = append(rules, allowAccessRule)
	// All other resources can only be accessed by admins.
	default:
		if claims.Admin() {
			rules = append(rules, allowAccessRule)
		}
	}
	return rules
}

// builtInAlertSecurityRule returns a built-in security rule to apply to an alert.
//
// TODO: This implementation is hard-coded specifically to properties currently set by the admin server.
// Should we refactor to a generic implementation where the admin server provides a conditional rule in the JWT instead?
func (p *securityEngine) builtInAlertSecurityRule(spec *runtimev1.AlertSpec, claims *SecurityClaims) *runtimev1.SecurityRule {
	// Allow if the user is an admin
	if claims.Admin() {
		return allowAccessRule
	}

	// Extract attributes
	var email, userID string
	if len(claims.UserAttributes) != 0 {
		userID, _ = claims.UserAttributes["id"].(string)
		email, _ = claims.UserAttributes["email"].(string)
	}

	// Allow if the owner is accessing the alert
	if spec.Annotations != nil && userID == spec.Annotations["admin_owner_user_id"] {
		return allowAccessRule
	}

	// Allow if the user is an email recipient
	for _, notifier := range spec.Notifiers {
		switch notifier.Connector {
		case "email":
			recipients := pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
			for _, recipient := range recipients {
				if recipient == email {
					return allowAccessRule
				}
			}
		case "slack":
			props, err := slack.DecodeProps(notifier.Properties.AsMap())
			if err != nil {
				p.logger.Error("failed to decode slack notifier properties", zap.Error(err))
				continue
			}
			for _, user := range props.Users {
				if user == email {
					return allowAccessRule
				}
			}
			// Note - A hack to allow slack channel users to access the alert. This also means that any alert configured with a slack channel will be viewable by any user part of the project and will appear in their alert list.
			if len(props.Channels) > 0 {
				return allowAccessRule
			}
		}
	}

	// Don't allow (but don't deny either)
	return nil
}

// builtInReportSecurityRule returns a built-in security rule to apply to a report.
//
// TODO: This implementation is hard-coded specifically to properties currently set by the admin server.
// Should we refactor to a generic implementation where the admin server provides a conditional rule in the JWT instead?
func (p *securityEngine) builtInReportSecurityRule(spec *runtimev1.ReportSpec, claims *SecurityClaims) *runtimev1.SecurityRule {
	// Allow if the user is an admin
	if claims.Admin() {
		return allowAccessRule
	}

	// Extract attributes
	var email, userID string
	if len(claims.UserAttributes) != 0 {
		userID, _ = claims.UserAttributes["id"].(string)
		email, _ = claims.UserAttributes["email"].(string)
	}

	// Allow if the owner is accessing the report
	if spec.Annotations != nil && userID == spec.Annotations["admin_owner_user_id"] {
		return allowAccessRule
	}

	// Allow if the user is an email recipient
	for _, notifier := range spec.Notifiers {
		switch notifier.Connector {
		case "email":
			recipients := pbutil.ToSliceString(notifier.Properties.AsMap()["recipients"])
			for _, recipient := range recipients {
				if recipient == email {
					return allowAccessRule
				}
			}
		case "slack":
			props, err := slack.DecodeProps(notifier.Properties.AsMap())
			if err != nil {
				p.logger.Error("failed to decode slack notifier properties", zap.Error(err))
				continue
			}
			for _, user := range props.Users {
				if user == email {
					return allowAccessRule
				}
			}
		}
	}

	// Don't allow (but don't deny either)
	return nil
}

// applySecurityRuleAccess applies an access rule to the resolved security.
func (p *securityEngine) applySecurityRuleAccess(res *ResolvedSecurity, r *runtimev1.Resource, rule *runtimev1.SecurityRuleAccess, td parser.TemplateData) error {
	// If already explicitly denied, do nothing (explicit denies take precedence over explicit allows)
	if res.access != nil && !*res.access {
		return nil
	}

	apply, err := evaluateConditions(r, rule.ConditionExpression, rule.ConditionKinds, rule.ConditionResources, td)
	if err != nil {
		return err
	}

	if apply == nil {
		// no conditions are provided
		res.access = &rule.Allow
		return nil
	}

	// Determine final access value
	allow := rule.Allow
	if rule.Exclusive {
		// Exclusive rules: apply the opposite when conditions don't match
		if !*apply {
			allow = !allow
		}
	} else if !*apply { // Non-exclusive rules: do nothing when conditions don't match
		return nil
	}

	res.access = &allow
	return nil
}

// applySecurityRuleFieldAccess applies a field access rule to the resolved security.
func (p *securityEngine) applySecurityRuleFieldAccess(res *ResolvedSecurity, r *runtimev1.Resource, rule *runtimev1.SecurityRuleFieldAccess, td parser.TemplateData) error {
	// This rule currently only applies to metrics views and explores.
	// Skip it for other resource types.
	var availableFields []string
	switch r.Meta.Name.Kind {
	case ResourceKindMetricsView:
		mv := r.GetMetricsView().State.ValidSpec
		if mv == nil {
			return nil
		}
		availableFields = make([]string, 0, len(mv.Dimensions)+len(mv.Measures))
		for _, f := range mv.Dimensions {
			availableFields = append(availableFields, f.Name)
		}
		for _, f := range mv.Measures {
			availableFields = append(availableFields, f.Name)
		}
	case ResourceKindExplore:
		exp := r.GetExplore().State.ValidSpec
		if exp == nil {
			return nil
		}
		availableFields = make([]string, 0, len(exp.Dimensions)+len(exp.Measures))
		availableFields = append(availableFields, exp.Dimensions...)
		availableFields = append(availableFields, exp.Measures...)
	default:
		return nil
	}

	// As soon as we see a field access rule, we set fieldAccess, which entails an implicit deny for all fields not mentioned
	if res.fieldAccess == nil {
		res.fieldAccess = make(map[string]bool)
	}

	apply, err := evaluateConditions(r, rule.ConditionExpression, rule.ConditionKinds, rule.ConditionResources, td)
	if err != nil {
		return err
	}
	if apply == nil || !*apply {
		return nil
	}

	// Set if the field should be allowed or denied
	if rule.AllFields {
		for _, f := range availableFields {
			v, ok := res.fieldAccess[f]
			if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
				res.fieldAccess[f] = rule.Allow
			}
		}
	} else {
		for _, f := range rule.Fields {
			v, ok := res.fieldAccess[f]
			if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
				res.fieldAccess[f] = rule.Allow
			}
		}
	}

	return nil
}

// applySecurityRuleRowFilter applies a row filter rule to the resolved security.
func (p *securityEngine) applySecurityRuleRowFilter(res *ResolvedSecurity, r *runtimev1.Resource, rule *runtimev1.SecurityRuleRowFilter, td parser.TemplateData) error {
	// Determine if the rule should be applied
	apply, err := evaluateConditions(r, rule.ConditionExpression, rule.ConditionKinds, rule.ConditionResources, td)
	if err != nil {
		return err
	}
	if apply == nil || !*apply {
		return nil
	}

	// Handle raw SQL row filters
	if rule.Sql != "" {
		sql, err := parser.ResolveTemplate(rule.Sql, td, false)
		if err != nil {
			return err
		}

		if res.rowFilter == "" {
			res.rowFilter = sql
		} else {
			res.rowFilter = fmt.Sprintf("(%s) AND (%s)", res.rowFilter, sql)
		}
	}

	// Handle query expression filters
	if rule.Expression != nil {
		if res.queryFilter == nil {
			res.queryFilter = rule.Expression
		} else {
			res.queryFilter = expressionpb.AndAll(res.queryFilter, rule.Expression)
		}
	}

	return nil
}

// expandTransitiveAccessRules expands any transitive access rules in the provided list of rules.
// This involves looking up the referenced resource, determining its dependencies, and adding the necessary access rules for those dependencies.
// For example, a transitive access rule on a report will add access rules for the underlying metrics view, explore, and any fields or rows that are accessible in the report.
func (p *securityEngine) expandTransitiveAccessRules(ctx context.Context, instanceID string, claims *SecurityClaims) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	// gather all conditions kinds and resources mentioned in the rules so that we can add a single security rule access policy for them at the end.
	// making sure only a single rule with the exclusive flag is added, otherwise we may get false rejections depending on which rule with the exclusive flag is evaluated first
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName
	for _, rule := range claims.AdditionalRules {
		if rule.GetTransitiveAccess() == nil {
			rules = append(rules, rule)
			continue
		}
		// If the rule is a transitive access rule, we need to resolve it
		resName := rule.GetTransitiveAccess().GetResource()
		if resName == nil {
			return nil, fmt.Errorf("transitive access rule has no resource")
		}
		ctr, err := p.rt.Controller(ctx, instanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get controller: %w", err)
		}
		res, err := ctr.Get(ctx, resName, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource %q of kind %q: %w", resName.Name, resName.Kind, err)
		}
		switch res.GetResource().(type) {
		case *runtimev1.Resource_Report:
			resolvedRules, ck, cr, err := p.resolveTransitiveAccessRuleForReport(ctx, instanceID, claims, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for report: %w", err)
			}
			rules = append(rules, resolvedRules...)
			conditionKinds = append(conditionKinds, ck...)
			conditionResources = append(conditionResources, cr...)
		case *runtimev1.Resource_Alert:
			resolvedRules, ck, cr, err := p.resolveTransitiveAccessRuleForAlert(ctx, instanceID, claims, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for alert: %w", err)
			}
			rules = append(rules, resolvedRules...)
			conditionKinds = append(conditionKinds, ck...)
			conditionResources = append(conditionResources, cr...)
		case *runtimev1.Resource_Explore:
			ck, cr, err := p.resolveTransitiveAccessRuleForExplore(res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for explore: %w", err)
			}
			conditionKinds = append(conditionKinds, ck...)
			conditionResources = append(conditionResources, cr...)
		case *runtimev1.Resource_Canvas:
			resolvedRules, ck, cr, err := p.resolveTransitiveAccessRuleForCanvas(ctx, instanceID, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for canvas: %w", err)
			}
			rules = append(rules, resolvedRules...)
			conditionKinds = append(conditionKinds, ck...)
			conditionResources = append(conditionResources, cr...)
		default:
			return nil, fmt.Errorf("transitive access rule for resource kind %q is not supported", res.Meta.Name.Kind)
		}
	}
	// add a security rule access policy for the resources mentioned in the conditions with exclusive flag set so that access is denied to everything else.
	if len(conditionKinds) > 0 || len(conditionResources) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionKinds:     conditionKinds,
					ConditionResources: conditionResources,
					Allow:              true,
					Exclusive:          true,
				},
			},
		})
	}

	return rules, nil
}

// resolveTransitiveAccessRuleForReport resolves transitive access rules for a report resource.
// This determines all the resources needed to access the report like the underlying metrics view, explore, etc. and adds the corresponding security rules to the list of rules to be applied.
// Also use the underlying query to determine the fields that are accessible in the report and where clause that needs to be applied.
// Restricts access to these fields and rows by adding corresponding field access and row filter rules to the resolved security rules.
func (p *securityEngine) resolveTransitiveAccessRuleForReport(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, []string, []*runtimev1.ResourceName, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionRes []*runtimev1.ResourceName

	report := res.GetReport()
	if report == nil {
		return nil, nil, nil, fmt.Errorf("resource is not a report")
	}

	spec := report.GetSpec()
	if spec == nil {
		return nil, nil, nil, fmt.Errorf("report spec is nil")
	}
	conditionRes = append(conditionRes, res.Meta.Name)
	conditionKinds = append(conditionKinds, ResourceKindTheme)

	if spec.QueryName != "" {
		initializer, ok := ResolverInitializers["legacy_metrics"]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no resolver found for name 'legacy_metrics'")
		}
		resolver, err := initializer(ctx, &ResolverOptions{
			Runtime:    p.rt,
			InstanceID: instanceID,
			Properties: map[string]any{
				"query_name":      spec.QueryName,
				"query_args_json": spec.QueryArgsJson,
			},
			Claims:    claims,
			ForExport: false,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, nil, nil, err
		}
		rules = append(rules, inferred...)

		mvName := ""
		refs := resolver.Refs()
		for _, ref := range refs {
			// need access to the referenced resources
			conditionRes = append(conditionRes, &runtimev1.ResourceName{Kind: ref.Kind, Name: ref.Name})
			if ref.Kind == ResourceKindMetricsView {
				mvName = ref.Name
			}
		}

		// figure out explore or canvas for the report
		var explore, canvas string
		if e, ok := spec.Annotations["explore"]; ok {
			explore = e
		}
		if c, ok := spec.Annotations["canvas"]; ok {
			canvas = c
		}

		if explore == "" { // backwards compatibility, try to find explore
			if path, ok := spec.Annotations["web_open_path"]; ok {
				// parse path, extract explore name, it will be like /explore/{explore}
				if strings.HasPrefix(path, "/explore/") {
					explore = path[9:]
					if explore[len(explore)-1] == '/' {
						explore = explore[:len(explore)-1]
					}
				}
			}
			// still not found, use mv name as explore name
			if explore == "" {
				explore = mvName
			}
		}

		if explore != "" {
			conditionRes = append(conditionRes, &runtimev1.ResourceName{Kind: ResourceKindExplore, Name: explore})
		}
		if canvas != "" {
			conditionRes = append(conditionRes, &runtimev1.ResourceName{Kind: ResourceKindCanvas, Name: canvas})
		}
	}

	return rules, conditionKinds, conditionRes, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForAlert(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, []string, []*runtimev1.ResourceName, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName

	alert := res.GetAlert()
	if alert == nil {
		return nil, nil, nil, fmt.Errorf("resource is not an alert")
	}

	spec := alert.GetSpec()
	if spec == nil {
		return nil, nil, nil, fmt.Errorf("alert spec is nil")
	}

	// explicitly allow access to the alert itself
	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, ResourceKindTheme)

	var mvName string
	if spec.QueryName != "" {
		initializer, ok := ResolverInitializers["legacy_metrics"]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no resolver found for name 'legacy_metrics'")
		}
		resolver, err := initializer(ctx, &ResolverOptions{
			Runtime:    p.rt,
			InstanceID: instanceID,
			Properties: map[string]any{
				"query_name":      spec.QueryName,
				"query_args_json": spec.QueryArgsJson,
			},
			Claims:    claims,
			ForExport: false,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, nil, nil, err
		}
		rules = append(rules, inferred...)

		refs := resolver.Refs()
		for _, ref := range refs {
			conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ref.Kind, Name: ref.Name})
		}
	}

	if spec.Resolver != "" {
		initializer, ok := ResolverInitializers[spec.Resolver]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no resolver found for name %q", spec.Resolver)
		}
		resolver, err := initializer(ctx, &ResolverOptions{
			Runtime:    p.rt,
			InstanceID: instanceID,
			Properties: spec.ResolverProperties.AsMap(),
			Claims:     claims,
			ForExport:  false,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, nil, nil, err
		}
		rules = append(rules, inferred...)

		refs := resolver.Refs()
		for _, ref := range refs {
			conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ref.Kind, Name: ref.Name})
		}
	}

	// figure out explore or canvas for the alert
	var explore, canvas string
	if e, ok := spec.Annotations["explore"]; ok {
		explore = e
	}
	if c, ok := spec.Annotations["canvas"]; ok {
		canvas = c
	}

	if explore == "" { // backwards compatibility, try to find explore
		if path, ok := spec.Annotations["web_open_path"]; ok {
			// parse path, extract explore name, it will be like /explore/{explore}
			if strings.HasPrefix(path, "/explore/") {
				explore = path[9:]
				if explore[len(explore)-1] == '/' {
					explore = explore[:len(explore)-1]
				}
			}
		}
		// still not found, use mv name as explore name // TODO does this harm anything? as some alerts may not have any explore like those based on sql resolvers
		if explore == "" {
			explore = mvName
		}
	}

	if explore != "" {
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ResourceKindExplore, Name: explore})
	}
	if canvas != "" {
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ResourceKindCanvas, Name: canvas})
	}

	return rules, conditionKinds, conditionResources, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForExplore(res *runtimev1.Resource) ([]string, []*runtimev1.ResourceName, error) {
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName

	explore := res.GetExplore()
	if explore == nil {
		return nil, nil, fmt.Errorf("resource is not an explore")
	}

	spec := explore.GetState().GetValidSpec()
	if spec == nil {
		return nil, nil, fmt.Errorf("explore valid spec is nil")
	}

	if spec.MetricsView == "" {
		return nil, nil, fmt.Errorf("explore does not reference a metrics view")
	}

	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, ResourceKindTheme)

	// give access to the underlying metrics view
	if spec.MetricsView != "" {
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: spec.MetricsView})
	}

	return conditionKinds, conditionResources, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForCanvas(ctx context.Context, instanceID string, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, []string, []*runtimev1.ResourceName, error) {
	var rules []*runtimev1.SecurityRule
	var conditionKinds []string
	var conditionResources []*runtimev1.ResourceName
	refs := &rendererRefs{
		metricsViews: make(map[string]bool),
		mvFields:     make(map[string]map[string]bool),
		mvFilters:    make(map[string][]string),
	}

	canvas := res.GetCanvas()
	if canvas == nil {
		return nil, nil, nil, fmt.Errorf("resource is not a canvas")
	}

	spec := canvas.GetState().GetValidSpec()
	if spec == nil {
		spec = canvas.GetSpec() // Fallback to spec if ValidSpec is not available
	}
	if spec == nil {
		return nil, nil, nil, fmt.Errorf("canvas spec is nil")
	}

	// explicitly allow access to the canvas itself
	conditionResources = append(conditionResources, res.Meta.Name)
	conditionKinds = append(conditionKinds, ResourceKindTheme)

	// Get controller to fetch components
	ctr, err := p.rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get controller: %w", err)
	}

	// Collect all component names referenced by the canvas
	componentNames := make(map[string]bool)
	for _, row := range spec.Rows {
		for _, item := range row.Items {
			componentNames[item.Component] = true
		}
	}

	// Process each component
	for componentName := range componentNames {
		componentRef := &runtimev1.ResourceName{
			Kind: ResourceKindComponent,
			Name: componentName,
		}
		// Allow access to the component itself
		conditionResources = append(conditionResources, componentRef)

		// Get component resource
		componentRes, err := ctr.Get(ctx, componentRef, false)
		if err != nil {
			// If component is not found, skip it but still allow access to the component name
			continue
		}

		// Get component spec to extract renderer properties
		componentSpec := componentRes.GetComponent().State.ValidSpec
		if componentSpec == nil {
			componentSpec = componentRes.GetComponent().Spec
		}

		if componentSpec.RendererProperties == nil {
			continue
		}

		rendererProps := componentSpec.RendererProperties.AsMap()
		err = populateRendererRefs(refs, componentSpec.Renderer, rendererProps)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse renderer properties for component %q: %w", componentName, err)
		}
	}

	// Now build security rules based on the collected references
	// First, allow access to all referenced metrics views
	// Then, for each metrics view, add field access and row filter rules as needed
	for mv := range refs.metricsViews {
		// allow access to the referenced metrics view
		conditionResources = append(conditionResources, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: mv})

		mvf, ok := refs.mvFields[mv]
		if ok && len(mvf) > 0 {
			fields := make([]string, 0, len(mvf))
			for f := range mvf {
				fields = append(fields, f)
			}
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_FieldAccess{
					FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						ConditionResources: []*runtimev1.ResourceName{{Kind: ResourceKindMetricsView, Name: mv}},
						Fields:             fields,
						Allow:              true,
					},
				},
			})
		}

		mvr, ok := refs.mvFilters[mv]
		if ok && len(mvr) > 0 {
			// Combine multiple row filters with OR
			rowFilter := strings.Join(mvr, " OR ")
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_RowFilter{
					RowFilter: &runtimev1.SecurityRuleRowFilter{
						ConditionResources: []*runtimev1.ResourceName{{Kind: ResourceKindMetricsView, Name: mv}},
						Sql:                rowFilter,
					},
				},
			})
		}
	}

	return rules, conditionKinds, conditionResources, nil
}

func evaluateConditions(r *runtimev1.Resource, expression string, kinds []string, resources []*runtimev1.ResourceName, td parser.TemplateData) (*bool, error) {
	// Evaluate resource-based conditions
	var resourceMatches *bool
	if len(kinds) > 0 || len(resources) > 0 {
		matches := slices.Contains(kinds, r.Meta.Name.Kind) ||
			slices.ContainsFunc(resources, func(res *runtimev1.ResourceName) bool {
				return res.Kind == r.Meta.Name.Kind && res.Name == r.Meta.Name.Name
			})
		resourceMatches = &matches
	}

	// Evaluate expression-based conditions
	var expressionMatches *bool
	if expression != "" {
		expr, err := parser.ResolveTemplate(expression, td, false)
		if err != nil {
			return nil, err
		}
		matches, err := parser.EvaluateBoolExpression(expr)
		if err != nil {
			return nil, err
		}
		expressionMatches = &matches
	}

	if resourceMatches == nil && expressionMatches == nil {
		// No conditions to evaluate
		return nil, nil
	}

	// Combine conditions (both must be true if both are present)
	conditionsMatch := true
	if resourceMatches != nil {
		conditionsMatch = *resourceMatches
	}
	if expressionMatches != nil {
		conditionsMatch = conditionsMatch && *expressionMatches
	}

	return &conditionsMatch, nil
}

// computeCacheKey computes a cache key for a resolved security policy.
func computeCacheKey(instanceID, environment string, claims *SecurityClaims, r *runtimev1.Resource) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(instanceID))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(environment))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(r.Meta.Name.Name))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(r.Meta.StateUpdatedOn.AsTime().String()))
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	_, err = hash.Write(claimsJSON)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// populateRendererRefs extracts all metricsview and its field names and filters from renderer properties based on the renderer type
// Depending on the component, fields will be named differently - Also there can be computed time dimension like <time_dim>_rill_TIME_GRAIN_<GRAIN>
//
//		"leaderboard" - "dimensions" and "measures"
//		"kpi_grid" - "dimensions" and "measures"
//		"table" - "columns" (can have computed time dim)
//		"pivot" - "row_dimensions", "col_dimensions" and "measures" (row/col can have computed time dim)
//		"heatmap" - "color"."field", "x"."field" and "y"."field"
//	 	"multi_metric_chart" - "measures" and "x"."field"
//		"funnel_chart" - "stage"."field", "measure"."field"
//		"donut_chart" - "color"."field", "measure"."field"
//		"bar_chart" - "color"."field", "x"."field" and "y"."field"
//		"line_chart" - "color"."field", "x"."field" and "y"."field"
//		"area_chart" - "color"."field", "x"."field" and "y"."field"
//		"stacked_bar" - "color"."field", "x"."field" and "y"."field"
//		"stacked_bar_normalized" - "color"."field", "x"."field" and "y"."field"
func populateRendererRefs(res *rendererRefs, renderer string, rendererProps map[string]any) error {
	mv, ok := pathutil.GetPath(rendererProps, "metrics_view")
	if !ok {
		return nil
	}
	res.metricsView(mv)
	filter, ok := pathutil.GetPath(rendererProps, "dimension_filters")
	if ok {
		res.metricsViewRowFilter(mv, filter)
	}
	switch renderer {
	case "leaderboard":
		dims, ok := pathutil.GetPath(rendererProps, "dimensions")
		if ok {
			res.metricsViewFields(mv, dims)
		}
		meas, ok := pathutil.GetPath(rendererProps, "measures")
		if ok {
			res.metricsViewFields(mv, meas)
		}
	case "kpi_grid":
		dims, ok := pathutil.GetPath(rendererProps, "dimensions")
		if ok {
			res.metricsViewFields(mv, dims)
		}
		meas, ok := pathutil.GetPath(rendererProps, "measures")
		if ok {
			res.metricsViewFields(mv, meas)
		}
	case "table":
		cols, ok := pathutil.GetPath(rendererProps, "columns")
		if ok {
			res.metricsViewFields(mv, cols)
		}
	case "pivot":
		rowDims, ok := pathutil.GetPath(rendererProps, "row_dimensions")
		if ok {
			res.metricsViewFields(mv, rowDims)
		}
		colDims, ok := pathutil.GetPath(rendererProps, "col_dimensions")
		if ok {
			res.metricsViewFields(mv, colDims)
		}
		meas, ok := pathutil.GetPath(rendererProps, "measures")
		if ok {
			res.metricsViewFields(mv, meas)
		}
	case "heatmap":
		colorField, ok := pathutil.GetPath(rendererProps, "color.field")
		if ok {
			res.metricsViewField(mv, colorField)
		}
		xField, ok := pathutil.GetPath(rendererProps, "x.field")
		if ok {
			res.metricsViewField(mv, xField)
		}
		yField, ok := pathutil.GetPath(rendererProps, "y.field")
		if ok {
			res.metricsViewField(mv, yField)
		}
	case "multi_metric_chart":
		meas, ok := pathutil.GetPath(rendererProps, "measures")
		if ok {
			res.metricsViewFields(mv, meas)
		}
		xField, ok := pathutil.GetPath(rendererProps, "x.field")
		if ok {
			res.metricsViewField(mv, xField)
		}
	case "funnel_chart":
		stageField, ok := pathutil.GetPath(rendererProps, "stage.field")
		if ok {
			res.metricsViewField(mv, stageField)
		}
		measureField, ok := pathutil.GetPath(rendererProps, "measure.field")
		if ok {
			res.metricsViewField(mv, measureField)
		}
	case "donut_chart":
		colorField, ok := pathutil.GetPath(rendererProps, "color.field")
		if ok {
			res.metricsViewField(mv, colorField)
		}
		measureField, ok := pathutil.GetPath(rendererProps, "measure.field")
		if ok {
			res.metricsViewField(mv, measureField)
		}
	case "bar_chart", "line_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		colorField, ok := pathutil.GetPath(rendererProps, "color.field")
		if ok {
			res.metricsViewField(mv, colorField)
		}
		xField, ok := pathutil.GetPath(rendererProps, "x.field")
		if ok {
			res.metricsViewField(mv, xField)
		}
		yField, ok := pathutil.GetPath(rendererProps, "y.field")
		if ok {
			res.metricsViewField(mv, yField)
		}
	default:
		return fmt.Errorf("unknown renderer type %q", renderer)
	}
	return nil
}

// extractDimension return the dimension or extracts the base time dimension from computed time field if present
// example - from "<time_dim>_rill_TIME_GRAIN_<GRAIN>" extracts "<time_dim>"
func extractDimension(field string) string {
	if strings.Contains(field, "_rill_TIME_GRAIN_") {
		parts := strings.Split(field, "_rill_TIME_GRAIN_")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return field
}

// extractFieldsFromSQLFilter parses a SQL filter and extracts field names directly during parsing
// This is a simplified version of metricssqlparser.ParseSQLFilter that only collects field names to avoid circular dependency issues
func extractFieldsFromSQLFilter(sqlFilter string) []string {
	if sqlFilter == "" {
		return nil
	}

	p := tidbparser.New()
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)
	sql := "SELECT * FROM tbl WHERE " + sqlFilter
	stmtNodes, _, err := p.ParseSQL(sql)
	if err != nil {
		return nil
	}

	if len(stmtNodes) != 1 {
		return nil
	}

	stmt, ok := stmtNodes[0].(*ast.SelectStmt)
	if !ok {
		return nil
	}

	fields := make(map[string]bool)
	extractFieldsFromNode(stmt.Where, fields)

	var result []string
	for field := range fields {
		if field != "" {
			result = append(result, field)
		}
	}

	return result
}

// extractFieldsFromNode recursively extracts field names from AST nodes
func extractFieldsFromNode(node ast.Node, fields map[string]bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.ColumnNameExpr:
		if n.Name != nil && n.Name.Schema.String() == "" && n.Name.Table.String() == "" {
			fields[n.Name.Name.O] = true
		}
	case *ast.BinaryOperationExpr:
		extractFieldsFromNode(n.L, fields)
		extractFieldsFromNode(n.R, fields)
	case *ast.IsNullExpr:
		extractFieldsFromNode(n.Expr, fields)
	case *ast.IsTruthExpr:
		extractFieldsFromNode(n.Expr, fields)
	case *ast.ParenthesesExpr:
		extractFieldsFromNode(n.Expr, fields)
	case *ast.PatternInExpr:
		extractFieldsFromNode(n.Expr, fields)
		for _, expr := range n.List {
			extractFieldsFromNode(expr, fields)
		}
	case *ast.PatternLikeOrIlikeExpr:
		extractFieldsFromNode(n.Expr, fields)
		extractFieldsFromNode(n.Pattern, fields)
	case *ast.BetweenExpr:
		extractFieldsFromNode(n.Expr, fields)
		extractFieldsFromNode(n.Left, fields)
		extractFieldsFromNode(n.Right, fields)
	case *ast.FuncCallExpr:
		for _, arg := range n.Args {
			extractFieldsFromNode(arg, fields)
		}
	}
}

type rendererRefs struct {
	metricsViews map[string]bool
	mvFields     map[string]map[string]bool
	mvFilters    map[string][]string
}

func (r *rendererRefs) metricsView(mv any) {
	if mv, ok := mv.(string); ok {
		r.metricsViews[mv] = true
	}
}

func (r *rendererRefs) metricsViewFields(mv, fields any) {
	metricsView, ok1 := mv.(string)
	fs, ok2 := fields.([]interface{})
	if ok1 && ok2 {
		if r.mvFields[metricsView] == nil {
			r.mvFields[metricsView] = make(map[string]bool)
		}
		for _, f := range fs {
			fstr, ok := f.(string)
			if !ok {
				panic("field is not a string")
			}
			r.mvFields[metricsView][extractDimension(fstr)] = true
		}
	}
}

func (r *rendererRefs) metricsViewField(mv, field any) {
	metricsView, ok1 := mv.(string)
	f, ok2 := field.(string)
	if ok1 && ok2 && f != "" {
		if r.mvFields[metricsView] == nil {
			r.mvFields[metricsView] = make(map[string]bool)
		}
		r.mvFields[metricsView][extractDimension(f)] = true
	}
}

func (r *rendererRefs) metricsViewRowFilter(mv, filter any) {
	metricsView, ok1 := mv.(string)
	f, ok2 := filter.(string)
	if ok1 && ok2 && f != "" {
		r.mvFilters[metricsView] = append(r.mvFilters[metricsView], fmt.Sprintf("(%s)", f)) // wrap in () to ensure correct precedence when combining multiple filters with OR
	}
	// Extract fields from dimension_filters SQL expression
	dimFilterFields := extractFieldsFromSQLFilter(f)
	if r.mvFields[metricsView] == nil {
		r.mvFields[metricsView] = make(map[string]bool)
	}
	for _, f := range dimFilterFields {
		r.mvFields[metricsView][extractDimension(f)] = true
	}
}
