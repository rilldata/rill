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
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var ErrForbidden = errors.New("action not allowed")

type ResolvedMetricsViewSecurity struct {
	Access      bool
	FieldAccess map[string]bool
	RowFilter   string
	QueryFilter *runtimev1.Expression
}

func (r *ResolvedMetricsViewSecurity) CanAccessField(field string) bool {
	if r == nil {
		return true
	}

	if !r.Access {
		return false
	}

	// If there are no field access rules, all fields are allowed
	if r.FieldAccess == nil {
		return true
	}

	// If not explicitly allowed, it's an implicit deny
	return r.FieldAccess[field]
}

var openAccess = &ResolvedMetricsViewSecurity{
	Access:      true,
	FieldAccess: nil,
	RowFilter:   "",
	QueryFilter: nil,
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

func (p *securityEngine) resolveMetricsViewSecurity(instanceID, environment string, attributes map[string]any, rules []*runtimev1.SecurityRule, mv *runtimev1.Resource) (*ResolvedMetricsViewSecurity, error) {
	// Combine rules with the metrics view's rules
	spec := mv.GetMetricsView().State.ValidSpec
	if spec != nil {
		rules = append(rules, spec.SecurityRules...)
	}

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

	// If attributes is empty that means auth is disabled and no user context is available.
	// Since we are controlling the attributes we can safely return the open policy.
	// TODO: Make this more explicit!
	if len(attributes) == 0 {
		return openAccess, nil
	}

	cacheKey, err := computeCacheKey(instanceID, environment, attributes, rules, mv)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedMetricsViewSecurity), nil
	}

	templateData := rillv1.TemplateData{
		Environment: environment,
		User:        attributes,
		Self:        rillv1.TemplateResource{Meta: mv.Meta},
	}

	// Apply rules
	var access *bool
	var rowFilter string
	var queryFilter *runtimev1.Expression
	var fieldAccess map[string]bool
	for _, rule := range rules {
		if rule == nil {
			continue
		}

		switch r := rule.Rule.(type) {
		case *runtimev1.SecurityRule_Access:
			// Determine if the rule should be applied
			apply := true
			if r.Access.Condition != "" {
				expr, err := rillv1.ResolveTemplate(r.Access.Condition, templateData)
				if err != nil {
					return nil, err
				}
				res, err := rillv1.EvaluateBoolExpression(expr)
				if err != nil {
					return nil, err
				}
				apply = res
			}
			if !apply {
				continue
			}

			// Explicit denies take precedence
			if access == nil {
				access = &r.Access.Allow
			} else {
				tmp := *access && r.Access.Allow
				access = &tmp
			}
		case *runtimev1.SecurityRule_FieldAccess:
			// As soon as we see a field access rule, we set fieldAccess, which entails an implicit deny for all fields not mentioned
			if fieldAccess == nil {
				fieldAccess = make(map[string]bool)
			}

			// Determine if the rule should be applied
			apply := true
			if r.FieldAccess.Condition != "" {
				expr, err := rillv1.ResolveTemplate(r.FieldAccess.Condition, templateData)
				if err != nil {
					return nil, err
				}
				res, err := rillv1.EvaluateBoolExpression(expr)
				if err != nil {
					return nil, err
				}
				apply = res
			}
			if !apply {
				continue
			}

			// Set if the field should be allowed or denied
			if r.FieldAccess.AllFields {
				for _, f := range spec.Dimensions {
					v, ok := fieldAccess[f.Name]
					if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
						fieldAccess[f.Name] = r.FieldAccess.Allow
					}
				}
				for _, f := range spec.Measures {
					v, ok := fieldAccess[f.Name]
					if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
						fieldAccess[f.Name] = r.FieldAccess.Allow
					}
				}
			} else {
				for _, f := range r.FieldAccess.Fields {
					v, ok := fieldAccess[f]
					if !ok || v { // Only update if not already denied (because deny takes precedence over allow)
						fieldAccess[f] = r.FieldAccess.Allow
					}
				}
			}
		case *runtimev1.SecurityRule_RowFilter:
			// Determine if the rule should be applied
			apply := true
			if r.RowFilter.Condition != "" {
				expr, err := rillv1.ResolveTemplate(r.RowFilter.Condition, templateData)
				if err != nil {
					return nil, err
				}
				res, err := rillv1.EvaluateBoolExpression(expr)
				if err != nil {
					return nil, err
				}
				apply = res
			}
			if !apply {
				continue
			}

			// Handle raw SQL row filters
			if r.RowFilter.Sql != "" {
				sql, err := rillv1.ResolveTemplate(r.RowFilter.Sql, templateData)
				if err != nil {
					return nil, err
				}

				if rowFilter == "" {
					rowFilter = sql
				} else {
					rowFilter = fmt.Sprintf("(%s) AND (%s)", rowFilter, sql)
				}
			}

			// Handle query expression filters
			if r.RowFilter.Expression != nil {
				if queryFilter == nil {
					queryFilter = r.RowFilter.Expression
				} else {
					queryFilter = expressionpb.AndAll(queryFilter, r.RowFilter.Expression)
				}
			}
		}
	}

	resolved := &ResolvedMetricsViewSecurity{
		Access:      access != nil && *access, // Access defaults to false
		FieldAccess: fieldAccess,
		RowFilter:   rowFilter,
		QueryFilter: queryFilter,
	}

	p.cache.Add(cacheKey, resolved)

	return resolved, nil
}

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
