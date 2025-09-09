package runtime

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	tidbparser "github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
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

	expandedRules, err := p.expandRules(ctx, instanceID, claims)
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
func (p *securityEngine) applySecurityRuleAccess(res *ResolvedSecurity, _ *runtimev1.Resource, rule *runtimev1.SecurityRuleAccess, td parser.TemplateData) error {
	// If already explicitly denied, do nothing (explicit denies take precedence over explicit allows)
	if res.access != nil && !*res.access {
		return nil
	}

	if rule.Condition != "" {
		expr, err := parser.ResolveTemplate(rule.Condition, td, false)
		if err != nil {
			return err
		}
		apply, err := parser.EvaluateBoolExpression(expr)
		if err != nil {
			return err
		}

		if !apply {
			return nil
		}
	}

	res.access = &rule.Allow

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

	// Determine if the rule should be applied
	if rule.Condition != "" {
		expr, err := parser.ResolveTemplate(rule.Condition, td, false)
		if err != nil {
			return err
		}
		apply, err := parser.EvaluateBoolExpression(expr)
		if err != nil {
			return err
		}

		if !apply {
			return nil
		}
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
func (p *securityEngine) applySecurityRuleRowFilter(res *ResolvedSecurity, _ *runtimev1.Resource, rule *runtimev1.SecurityRuleRowFilter, td parser.TemplateData) error {
	// Determine if the rule should be applied
	if rule.Condition != "" {
		expr, err := parser.ResolveTemplate(rule.Condition, td, false)
		if err != nil {
			return err
		}
		apply, err := parser.EvaluateBoolExpression(expr)
		if err != nil {
			return err
		}

		if !apply {
			return nil
		}
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

func (p *securityEngine) expandRules(ctx context.Context, instanceID string, claims *SecurityClaims) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
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
			resolvedRules, err := p.resolveTransitiveAccessRuleForReport(ctx, instanceID, claims, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for report: %w", err)
			}
			rules = append(rules, resolvedRules...)
		case *runtimev1.Resource_Alert:
			resolvedRules, err := p.resolveTransitiveAccessRuleForAlert(ctx, instanceID, claims, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for alert: %w", err)
			}
			rules = append(rules, resolvedRules...)
		case *runtimev1.Resource_Explore:
			resolvedRules, err := p.resolveTransitiveAccessRuleForExplore(res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for explore: %w", err)
			}
			rules = append(rules, resolvedRules...)
		case *runtimev1.Resource_Canvas:
			resolvedRules, err := p.resolveTransitiveAccessRuleForCanvas(ctx, instanceID, res)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve transitive access rule for canvas: %w", err)
			}
			rules = append(rules, resolvedRules...)
		default:
			return nil, fmt.Errorf("transitive access rule for resource kind %q is not supported", res.Meta.Name.Kind)
		}
	}
	return rules, nil
}

// resolveTransitiveAccessRuleForReport resolves transitive access rules for a report resource.
// This determines all the resources needed to access the report like the underlying metrics view, explore, etc. and adds the corresponding security rules to the list of rules to be applied.
// Also use the underlying query to determine the fields that are accessible in the report and where clause that needs to be applied.
// Restricts access to these fields and rows by adding corresponding field access and row filter rules to the resolved security rules.
func (p *securityEngine) resolveTransitiveAccessRuleForReport(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule

	report := res.GetReport()
	if report == nil {
		return nil, fmt.Errorf("resource is not a report")
	}

	spec := report.GetSpec()
	if spec == nil {
		return nil, fmt.Errorf("report spec is nil")
	}
	// explicitly allow access to the report itself
	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindReport, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))),
				Allow:     true,
			},
		},
	})

	// deny everything except the report itself, themes, explore, canvas and metrics view
	var denyCondition strings.Builder
	// self report
	denyCondition.WriteString(fmt.Sprintf("('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindReport, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))))
	// all themes
	denyCondition.WriteString(fmt.Sprintf(" OR '{{.self.kind}}'='%s'", ResourceKindTheme))

	if spec.QueryName != "" {
		initializer, ok := ResolverInitializers["legacy_metrics"]
		if !ok {
			return nil, fmt.Errorf("no resolver found for name 'legacy_metrics'")
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
			return nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, err
		}
		rules = append(rules, inferred...)

		mvName := ""
		refs := resolver.Refs()
		for _, ref := range refs {
			// allow access to the referenced resource
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_Access{
					Access: &runtimev1.SecurityRuleAccess{
						Condition: fmt.Sprintf("'{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))),
						Allow:     true,
					},
				},
			})
			// add to deny condition
			denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s)", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))))
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
			denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindExplore, duckdbsql.EscapeStringValue(strings.ToLower(explore))))
		}
		if canvas != "" {
			denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindCanvas, duckdbsql.EscapeStringValue(strings.ToLower(canvas))))
		}
	}

	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("NOT (%s)", denyCondition.String()),
				Allow:     false,
			},
		},
	})

	return rules, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForAlert(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule

	alert := res.GetAlert()
	if alert == nil {
		return nil, fmt.Errorf("resource is not an alert")
	}

	spec := alert.GetSpec()
	if spec == nil {
		return nil, fmt.Errorf("alert spec is nil")
	}

	// explicitly allow access to the alert itself
	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindAlert, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))),
				Allow:     true,
			},
		},
	})

	var mvName string
	// deny everything except the alert itself and themes
	var denyCondition strings.Builder
	// self alert
	denyCondition.WriteString(fmt.Sprintf("('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindAlert, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))))
	// all themes
	denyCondition.WriteString(fmt.Sprintf(" OR '{{.self.kind}}'='%s'", ResourceKindTheme))

	if spec.QueryName != "" {
		initializer, ok := ResolverInitializers["legacy_metrics"]
		if !ok {
			return nil, fmt.Errorf("no resolver found for name 'legacy_metrics'")
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
			return nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, err
		}
		rules = append(rules, inferred...)

		refs := resolver.Refs()
		for _, ref := range refs {
			// allow access to the referenced resource
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_Access{
					Access: &runtimev1.SecurityRuleAccess{
						Condition: fmt.Sprintf("'{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))),
						Allow:     true,
					},
				},
			})
			// add to deny condition
			denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s)", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))))
			if ref.Kind == ResourceKindMetricsView {
				mvName = ref.Name
			}
		}
	}

	if spec.Resolver != "" {
		initializer, ok := ResolverInitializers[spec.Resolver]
		if !ok {
			return nil, fmt.Errorf("no resolver found for name %q", spec.Resolver)
		}
		resolver, err := initializer(ctx, &ResolverOptions{
			Runtime:    p.rt,
			InstanceID: instanceID,
			Properties: spec.ResolverProperties.AsMap(),
			Claims:     claims,
			ForExport:  false,
		})
		if err != nil {
			return nil, err
		}
		defer resolver.Close()
		inferred, err := resolver.InferRequiredSecurityRules()
		if err != nil {
			return nil, err
		}
		rules = append(rules, inferred...)

		refs := resolver.Refs()
		for _, ref := range refs {
			// allow access to the referenced resource
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_Access{
					Access: &runtimev1.SecurityRuleAccess{
						Condition: fmt.Sprintf("'{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))),
						Allow:     true,
					},
				},
			})
			// add to deny condition
			denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s)", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))))
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
		denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindExplore, duckdbsql.EscapeStringValue(strings.ToLower(explore))))
	}
	if canvas != "" {
		denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindCanvas, duckdbsql.EscapeStringValue(strings.ToLower(canvas))))
	}

	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("NOT (%s)", denyCondition.String()),
				Allow:     false,
			},
		},
	})

	return rules, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForExplore(res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule

	explore := res.GetExplore()
	if explore == nil {
		return nil, fmt.Errorf("resource is not an explore")
	}

	spec := explore.GetState().GetValidSpec()
	if spec == nil {
		return nil, fmt.Errorf("explore valid spec is nil")
	}

	if spec.MetricsView == "" {
		return nil, fmt.Errorf("explore does not reference a metrics view")
	}

	// explicitly allow access to the explore itself
	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindExplore, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))),
				Allow:     true,
			},
		},
	})

	// give access to the underlying metrics view
	if spec.MetricsView != "" {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindMetricsView, duckdbsql.EscapeStringValue(strings.ToLower(spec.MetricsView))),
					Allow:     true,
				},
			},
		})
	}

	// deny everything except the explore, mv and themes
	var denyCondition strings.Builder
	// self canvas
	denyCondition.WriteString(fmt.Sprintf("('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindExplore, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))))
	// underlying mv
	denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindMetricsView, duckdbsql.EscapeStringValue(strings.ToLower(spec.MetricsView))))
	// all themes
	denyCondition.WriteString(fmt.Sprintf(" OR '{{.self.kind}}'='%s'", ResourceKindTheme))

	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("NOT (%s)", denyCondition.String()),
				Allow:     false,
			},
		},
	})

	return rules, nil
}

func (p *securityEngine) resolveTransitiveAccessRuleForCanvas(ctx context.Context, instanceID string, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule

	canvas := res.GetCanvas()
	if canvas == nil {
		return nil, fmt.Errorf("resource is not a canvas")
	}

	spec := canvas.GetState().GetValidSpec()
	if spec == nil {
		spec = canvas.GetSpec() // Fallback to spec if ValidSpec is not available
	}
	if spec == nil {
		return nil, fmt.Errorf("canvas spec is nil")
	}

	// explicitly allow access to the canvas itself
	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindCanvas, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))),
				Allow:     true,
			},
		},
	})

	// deny everything except the canvas itself, themes, and components
	var denyCondition strings.Builder
	// self canvas
	denyCondition.WriteString(fmt.Sprintf("('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindCanvas, duckdbsql.EscapeStringValue(strings.ToLower(res.Meta.Name.Name))))
	// all themes
	denyCondition.WriteString(fmt.Sprintf(" OR '{{.self.kind}}'='%s'", ResourceKindTheme))

	// Get controller to fetch components
	ctr, err := p.rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}

	// Collect all component names referenced by the canvas
	componentNames := make(map[string]bool)
	for _, row := range spec.Rows {
		for _, item := range row.Items {
			componentNames[item.Component] = true
		}
	}

	// for deduplicating and combing fields for metrics view referenced by multiple components
	mvSecInfoCache := make(map[string]*metricsViewSecurityInfo)
	seenRefs := make(map[string]bool) // optional optimization to avoid adding access rule and deny condition multiple times for the same resource

	// Process each component
	for componentName := range componentNames {
		// Allow access to the component itself
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindComponent, duckdbsql.EscapeStringValue(strings.ToLower(componentName))),
					Allow:     true,
				},
			},
		})

		// Add to deny condition
		denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s)", ResourceKindComponent, duckdbsql.EscapeStringValue(strings.ToLower(componentName))))

		// Get component resource
		componentRef := &runtimev1.ResourceName{
			Kind: ResourceKindComponent,
			Name: componentName,
		}
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

		// Allow access to metrics views referenced by this component and create field/row filter rules
		for _, ref := range componentRes.Meta.Refs {
			_, seen := seenRefs[fmt.Sprintf("%s:%s", ref.Kind, strings.ToLower(ref.Name))]
			if !seen {
				// allow access to the referenced resource
				rules = append(rules, &runtimev1.SecurityRule{
					Rule: &runtimev1.SecurityRule_Access{
						Access: &runtimev1.SecurityRuleAccess{
							Condition: fmt.Sprintf("'{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))),
							Allow:     true,
						},
					},
				})
				// add to deny condition
				denyCondition.WriteString(fmt.Sprintf(" OR ('{{.self.kind}}'=%s AND '{{lower .self.name}}'=%s)", duckdbsql.EscapeStringValue(ref.Kind), duckdbsql.EscapeStringValue(strings.ToLower(ref.Name))))
			}
			seenRefs[fmt.Sprintf("%s:%s", ref.Kind, strings.ToLower(ref.Name))] = true

			// handle metrics view
			if ref.Kind == ResourceKindMetricsView && componentSpec.RendererProperties != nil {
				mvSecInfo, exists := mvSecInfoCache[ref.Name]
				if !exists {
					mvSecInfo = &metricsViewSecurityInfo{fieldSet: make(map[string]bool)}
					mvSecInfoCache[ref.Name] = mvSecInfo
				}

				// Extract renderer properties
				rendererProps := componentSpec.RendererProperties.AsMap()

				// Extract fields using the comprehensive helper method that handles all renderer types
				fields, err := extractFieldsFromRendererProperties(componentSpec.Renderer, rendererProps)
				if err != nil {
					return nil, fmt.Errorf("failed to extract fields from renderer properties for component %q: %w", componentName, err)
				}
				for _, field := range fields {
					mvSecInfo.fieldSet[field] = true
				}

				// Parse dimension_filters and create row filter rule
				if dimFilter, ok := rendererProps["dimension_filters"].(string); ok && dimFilter != "" {
					mvSecInfo.rowFilters = append(mvSecInfo.rowFilters, fmt.Sprintf("(%s)", dimFilter))

					// Extract fields from dimension_filters SQL expression
					dimFilterFields := extractFieldsFromSQLFilter(dimFilter)
					for _, field := range dimFilterFields {
						mvSecInfo.fieldSet[field] = true
					}
				}
			}
		}
	}

	// add field access and row filter rules for each metrics view
	for mvName, mvSecInfo := range mvSecInfoCache {
		// Create field access rule if we have fields
		if len(mvSecInfo.fieldSet) > 0 {
			fields := make([]string, 0, len(mvSecInfo.fieldSet))
			for f := range mvSecInfo.fieldSet {
				fields = append(fields, f)
			}
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_FieldAccess{
					FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindMetricsView, duckdbsql.EscapeStringValue(strings.ToLower(mvName))),
						Fields:    fields,
						Allow:     true,
					},
				},
			})
		}

		// Create row filter rule if we have row filters
		if len(mvSecInfo.rowFilters) > 0 {
			// Combine multiple row filters with OR
			rowFilter := strings.Join(mvSecInfo.rowFilters, " OR ")
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_RowFilter{
					RowFilter: &runtimev1.SecurityRuleRowFilter{
						Condition: fmt.Sprintf("'{{.self.kind}}'='%s' AND '{{lower .self.name}}'=%s", ResourceKindMetricsView, duckdbsql.EscapeStringValue(strings.ToLower(mvName))),
						Sql:       rowFilter,
					},
				},
			})
		}
	}

	// Add deny rule for everything else
	rules = append(rules, &runtimev1.SecurityRule{
		Rule: &runtimev1.SecurityRule_Access{
			Access: &runtimev1.SecurityRuleAccess{
				Condition: fmt.Sprintf("NOT (%s)", denyCondition.String()),
				Allow:     false,
			},
		},
	})

	return rules, nil
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

// extractFieldsFromRendererProperties extracts field names from renderer properties based on the renderer type, can contain duplicate fields
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
func extractFieldsFromRendererProperties(renderer string, rendererProps map[string]any) ([]string, error) {
	switch renderer {
	case "leaderboard", "kpi_grid", "table", "pivot", "heatmap", "multi_metric_chart", "bar_chart", "line_chart", "area_chart", "stacked_bar", "stacked_bar_normalized", "funnel_chart", "donut_chart":
	default:
		return nil, fmt.Errorf("unknown renderer type %q", renderer)
	}

	var fields []string

	// Try common array fields
	arrayFields := []string{"dimensions", "measures", "columns", "row_dimensions", "col_dimensions"}
	for _, key := range arrayFields {
		if arr, ok := rendererProps[key]; ok {
			rawFields := extractFieldsFromArray(arr)
			for _, field := range rawFields {
				fields = append(fields, extractDimension(field))
			}
		}
	}

	// Try common nested object fields
	objectFields := []string{"color", "x", "y", "stage", "measure"}
	for _, key := range objectFields {
		if obj, ok := rendererProps[key].(map[string]interface{}); ok {
			if field, ok := obj["field"].(string); ok && field != "" {
				fields = append(fields, extractDimension(field))
			}
		}
	}

	return fields, nil
}

// extractFieldsFromArray extracts field names from string array
func extractFieldsFromArray(arr interface{}) []string {
	var fields []string
	if arrSlice, ok := arr.([]interface{}); ok {
		for _, item := range arrSlice {
			if str, ok := item.(string); ok {
				fields = append(fields, str)
			}
		}
	}
	return fields
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

type metricsViewSecurityInfo struct {
	fieldSet   map[string]bool
	rowFilters []string
}
