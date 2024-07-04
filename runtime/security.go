package runtime

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var ErrForbidden = errors.New("action not allowed")

type ResolvedSecurity struct {
	access      *bool
	fieldAccess map[string]bool
	rowFilter   string
	queryFilter *runtimev1.Expression
}

func (r *ResolvedSecurity) CanAccess() bool {
	if r == nil {
		return true
	}
	if r.access == nil {
		return false
	}
	return *r.access
}

func (r *ResolvedSecurity) CanAccessAllFields() bool {
	if r == nil {
		return true
	}
	// If there are no field access rules, all fields are allowed
	return r.fieldAccess == nil
}

func (r *ResolvedSecurity) CanAccessField(field string) bool {
	if r == nil {
		return true
	}

	if !r.CanAccess() {
		return false
	}

	if r.CanAccessAllFields() {
		return true
	}

	// If not explicitly allowed, it's an implicit deny
	return r.fieldAccess[field]
}

func (r *ResolvedSecurity) RowFilter() string {
	if r == nil {
		return ""
	}
	return r.rowFilter
}

func (r *ResolvedSecurity) QueryFilter() *runtimev1.Expression {
	if r == nil {
		return nil
	}
	return r.queryFilter
}

var truth = true

// openAccess is allows access to everything.
var openAccess = &ResolvedSecurity{
	access:      &truth,
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

// denyAccessRule is a security rule that denies access.
var denyAccessRule = &runtimev1.SecurityRule{
	Rule: &runtimev1.SecurityRule_Access{
		Access: &runtimev1.SecurityRuleAccess{
			Allow: false,
		},
	},
}

type securityEngine struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	logger *zap.Logger
}

func newSecurityEngine(cacheSize int, logger *zap.Logger) *securityEngine {
	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		panic(err)
	}
	return &securityEngine{cache: cache, logger: logger}
}

// resolveSecurity resolves the security rules for a given resource and user context.
func (p *securityEngine) resolveSecurity(instanceID, environment string, attributes map[string]any, rules []*runtimev1.SecurityRule, r *runtimev1.Resource) (*ResolvedSecurity, error) {
	// If attributes is empty that means auth is disabled and no user context is available.
	// Since we are controlling the attributes we can safely return the open policy.
	// TODO: Make this more explicit!
	if len(attributes) == 0 {
		return openAccess, nil
	}

	// Combine rules with any contained in the resource itself
	rules = p.resolveRules(attributes, rules, r)

	// Exit early if all rules are nil
	var validRule bool
	for _, rule := range rules {
		if rule != nil {
			validRule = true
			break
		}
	}
	if !validRule {
		return nil, nil
	}

	cacheKey, err := computeCacheKey(instanceID, environment, attributes, rules, r)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedSecurity), nil
	}

	templateData := rillv1.TemplateData{
		Environment: environment,
		User:        attributes,
		Self:        rillv1.TemplateResource{Meta: r.Meta},
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
				return nil, err
			}
		case *runtimev1.SecurityRule_FieldAccess:
			err := p.applySecurityRuleFieldAccess(res, r, rule.FieldAccess, templateData)
			if err != nil {
				return nil, err
			}
		case *runtimev1.SecurityRule_RowFilter:
			err := p.applySecurityRuleRowFilter(res, r, rule.RowFilter, templateData)
			if err != nil {
				return nil, err
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

// resolveRules combines the provided rules with hardcoded rules and rules declared in the resource itself.
func (p *securityEngine) resolveRules(attributes map[string]any, rules []*runtimev1.SecurityRule, r *runtimev1.Resource) []*runtimev1.SecurityRule {
	rule := p.builtInKindSecurityRule(r.Meta.Name.Kind, attributes)
	if rule != nil {
		// Optimization for the only return value as of this writing.
		// If the rule is deny, we don't need to check any other rules.
		if rule == denyAccessRule {
			return []*runtimev1.SecurityRule{rule}
		}

		// Prepend instead of append since the rule is likely to lead to a quick deny access
		rules = append([]*runtimev1.SecurityRule{rule}, rules...)
	}

	switch r.Meta.Name.Kind {
	case ResourceKindMetricsView:
		spec := r.GetMetricsView().State.ValidSpec
		if spec != nil {
			rules = append(rules, spec.SecurityRules...)
		}
	case ResourceKindAlert:
		spec := r.GetAlert().Spec
		rule := p.builtInAlertSecurityRule(spec, attributes)
		if rule != nil {
			// Prepend instead of append since the rule is likely to lead to a quick deny access
			rules = append([]*runtimev1.SecurityRule{rule}, rules...)
		}
	case ResourceKindReport:
		spec := r.GetReport().Spec
		rule := p.builtInReportSecurityRule(spec, attributes)
		if rule != nil {
			// Prepend instead of append since the rule is likely to lead to a quick deny access
			rules = append([]*runtimev1.SecurityRule{rule}, rules...)
		}
	}

	return rules
}

// builtInKindSecurityRule returns a built-in security rule that checks if the user is allowed to access the resource kind.
//
// TODO: This implementation is hard-coded specifically to properties currently set by the admin server.
// We should refactor to a generic implementation where the admin server provides a conditional rule in the JWT instead.
func (p *securityEngine) builtInKindSecurityRule(kind string, attributes map[string]any) *runtimev1.SecurityRule {
	// Determine if the user is an admin
	admin := true // If no attributes are set, we need to assume it's an admin (TODO: make explicit)
	if len(attributes) != 0 {
		admin, _ = attributes["admin"].(bool)
	}

	// Don't add a rule if the user is an admin
	if admin {
		return nil
	}

	// Add a deny rule for certain resource kinds that only admins should access
	switch kind {
	case ResourceKindSource, ResourceKindModel, ResourceKindMigration, ResourceKindConnector:
		return denyAccessRule
	}

	return nil
}

// builtInAlertSecurityRule returns a built-in security rule to apply to an alert.
//
// TODO: This implementation is hard-coded specifically to properties currently set by the admin server.
// We should refactor to a generic implementation where the admin server provides a conditional rule in the JWT instead.
func (p *securityEngine) builtInAlertSecurityRule(spec *runtimev1.AlertSpec, attributes map[string]any) *runtimev1.SecurityRule {
	// Extract attributes
	var email, userID string
	admin := true // If no attributes are set, we need to assume it's an admin (TODO: make explicit)
	if len(attributes) != 0 {
		userID, _ = attributes["sub"].(string)
		email, _ = attributes["email"].(string)
		admin, _ = attributes["admin"].(bool)
	}

	// Allow if the owner is accessing the alert
	if spec.Annotations != nil && userID == spec.Annotations["admin_owner_user_id"] {
		return allowAccessRule
	}

	// Allow if the user is an admin
	if admin {
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

// builtInReportSecurityRule returns a built-in security rule to apply to a report.
//
// TODO: This implementation is hard-coded specifically to properties currently set by the admin server.
// We should refactor to a generic implementation where the admin server provides a conditional rule in the JWT instead.
func (p *securityEngine) builtInReportSecurityRule(spec *runtimev1.ReportSpec, attributes map[string]any) *runtimev1.SecurityRule {
	// Extract attributes
	var email, userID string
	admin := true // If no attributes are set, we need to assume it's an admin (TODO: make explicit)
	if len(attributes) != 0 {
		userID, _ = attributes["sub"].(string)
		email, _ = attributes["email"].(string)
		admin, _ = attributes["admin"].(bool)
	}

	// Allow if the owner is accessing the report
	if spec.Annotations != nil && userID == spec.Annotations["admin_owner_user_id"] {
		return allowAccessRule
	}

	// Allow if the user is an admin
	if admin {
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
func (p *securityEngine) applySecurityRuleAccess(res *ResolvedSecurity, _ *runtimev1.Resource, rule *runtimev1.SecurityRuleAccess, td rillv1.TemplateData) error {
	// If already explicitly denied, do nothing (explicit denies take precedence over explicit allows)
	if res.access != nil && !*res.access {
		return nil
	}

	if rule.Condition != "" {
		expr, err := rillv1.ResolveTemplate(rule.Condition, td)
		if err != nil {
			return err
		}
		apply, err := rillv1.EvaluateBoolExpression(expr)
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
func (p *securityEngine) applySecurityRuleFieldAccess(res *ResolvedSecurity, r *runtimev1.Resource, rule *runtimev1.SecurityRuleFieldAccess, td rillv1.TemplateData) error {
	// This rule currently only applies to metrics views.
	// Skip it for other resource types.
	if r.Meta.Name.Kind != ResourceKindMetricsView {
		return nil
	}
	mv := r.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil
	}

	// As soon as we see a field access rule, we set fieldAccess, which entails an implicit deny for all fields not mentioned
	if res.fieldAccess == nil {
		res.fieldAccess = make(map[string]bool)
	}

	// Determine if the rule should be applied
	if rule.Condition != "" {
		expr, err := rillv1.ResolveTemplate(rule.Condition, td)
		if err != nil {
			return err
		}
		apply, err := rillv1.EvaluateBoolExpression(expr)
		if err != nil {
			return err
		}

		if !apply {
			return nil
		}
	}

	// Set if the field should be allowed or denied
	if rule.AllFields {
		for _, f := range mv.Dimensions {
			v, ok := res.fieldAccess[f.Name]
			if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
				res.fieldAccess[f.Name] = rule.Allow
			}
		}
		for _, f := range mv.Measures {
			v, ok := res.fieldAccess[f.Name]
			if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
				res.fieldAccess[f.Name] = rule.Allow
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
func (p *securityEngine) applySecurityRuleRowFilter(res *ResolvedSecurity, _ *runtimev1.Resource, rule *runtimev1.SecurityRuleRowFilter, td rillv1.TemplateData) error {
	// Determine if the rule should be applied
	if rule.Condition != "" {
		expr, err := rillv1.ResolveTemplate(rule.Condition, td)
		if err != nil {
			return err
		}
		apply, err := rillv1.EvaluateBoolExpression(expr)
		if err != nil {
			return err
		}

		if !apply {
			return nil
		}
	}

	// Handle raw SQL row filters
	if rule.Sql != "" {
		sql, err := rillv1.ResolveTemplate(rule.Sql, td)
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

// computeCacheKey computes a cache key for a resolved security policy.
func computeCacheKey(instanceID, environment string, attributes map[string]any, rules []*runtimev1.SecurityRule, r *runtimev1.Resource) (string, error) {
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
	// go through attributes in a deterministic order (alphabetical by keys)
	if attributes["admin"] != nil {
		err = binary.Write(hash, binary.BigEndian, attributes["admin"])
		if err != nil {
			return "", err
		}
	}
	if attributes["email"] != nil {
		_, err = hash.Write([]byte(attributes["email"].(string)))
		if err != nil {
			return "", err
		}
	}
	if attributes["groups"] != nil {
		for _, g := range attributes["groups"].([]interface{}) {
			_, err = hash.Write([]byte(g.(string)))
			if err != nil {
				return "", err
			}
		}
	}
	for _, r := range rules {
		if r == nil {
			continue
		}
		res, err := protojson.Marshal(r)
		if err != nil {
			return "", err
		}
		_, err = hash.Write(res)
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
