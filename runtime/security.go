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

var openAccess = &ResolvedMetricsViewSecurity{
	Access:      true,
	RowFilter:   "",
	QueryFilter: nil,
	Include:     nil,
	Exclude:     nil,
}

type ResolvedMetricsViewSecurity struct {
	Access      bool
	RowFilter   string
	QueryFilter *runtimev1.Expression
	Include     []string
	Exclude     []string
	ExcludeAll  bool
}

func (r *ResolvedMetricsViewSecurity) CanAccessField(field string) bool {
	if r == nil {
		return true
	}

	if !r.Access {
		return false
	}

	if r.ExcludeAll {
		return false
	}

	if len(r.Include) > 0 {
		for _, include := range r.Include {
			if include == field {
				return true
			}
		}
		return false
	}

	if len(r.Exclude) > 0 {
		for _, exclude := range r.Exclude {
			if exclude == field {
				return false
			}
		}
		return true
	}

	return true
}

func computeCacheKey(instanceID string, r *runtimev1.Resource, attributes map[string]any, policies ...*runtimev1.MetricsViewSpec_SecurityV2) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(instanceID))
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
	for _, p := range policies {
		if p == nil {
			continue
		}
		res, err := protojson.Marshal(p)
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

func (p *securityEngine) resolveMetricsViewSecurity(instanceID, environment string, attributes map[string]any, mv *runtimev1.Resource, policies ...*runtimev1.MetricsViewSpec_SecurityV2) (*ResolvedMetricsViewSecurity, error) {
	// Exit early if all policies are nil
	var validPolicy bool
	for _, policy := range policies {
		if policy != nil {
			validPolicy = true
			break
		}
	}
	if !validPolicy {
		return nil, nil
	}

	// if attributes is empty that means auth is disabled and also no user context is available
	// since we are controlling the attributes we can safely return the open policy
	if len(attributes) == 0 {
		return openAccess, nil
	}

	cacheKey, err := computeCacheKey(instanceID, mv, attributes, policies...)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedMetricsViewSecurity), nil
	}

	resolved := &ResolvedMetricsViewSecurity{Access: true}
	templateData := rillv1.TemplateData{
		Environment: environment,
		User:        attributes,
		Self:        rillv1.TemplateResource{Meta: mv.Meta},
	}

	for _, policy := range policies {
		if policy == nil {
			continue
		}

		if policy.Access != "" {
			accessExpr, err := rillv1.ResolveTemplate(policy.Access, templateData)
			if err != nil {
				return nil, err
			}
			access, err := rillv1.EvaluateBoolExpression(accessExpr)
			if err != nil {
				return nil, err
			}

			resolved.Access = resolved.Access && access
		} else {
			resolved.Access = false
		}

		if policy.RowFilter != "" {
			filter, err := rillv1.ResolveTemplate(policy.RowFilter, templateData)
			if err != nil {
				return nil, err
			}

			if resolved.RowFilter == "" {
				resolved.RowFilter = filter
			} else {
				resolved.RowFilter = fmt.Sprintf("(%s) AND (%s)", resolved.RowFilter, filter)
			}
		}

		if policy.QueryFilter != nil {
			if resolved.QueryFilter == nil {
				resolved.QueryFilter = policy.QueryFilter
			} else {
				expressionpb.And([]*runtimev1.Expression{resolved.QueryFilter, policy.QueryFilter})
			}
		}

		var include, exclude []string
		var excludeAll bool
		seen := map[string]bool{}

		for _, inc := range policy.Include {
			cond, err := rillv1.ResolveTemplate(inc.Condition, templateData)
			if err != nil {
				return nil, err
			}
			incCond, err := rillv1.EvaluateBoolExpression(cond)
			if err != nil {
				return nil, err
			}
			if incCond {
				for _, name := range inc.Names {
					if seen[name] {
						continue
					}
					seen[name] = true
					include = append(include, name)
				}
			}
		}

		// this is to handle the case where include filter was present but none of them evaluted to true
		if len(policy.Include) > 0 && len(include) == 0 {
			excludeAll = true
		}

		for _, exc := range policy.Exclude {
			cond, err := rillv1.ResolveTemplate(exc.Condition, templateData)
			if err != nil {
				return nil, err
			}
			excCond, err := rillv1.EvaluateBoolExpression(cond)
			if err != nil {
				return nil, err
			}
			if excCond {
				for _, name := range exc.Names {
					if seen[name] {
						continue
					}
					seen[name] = true
					exclude = append(exclude, name)
				}
			}
		}

		// Merge into resolved
		mergeIncludeExcludes(resolved, include, exclude, excludeAll)
	}

	p.cache.Add(cacheKey, resolved)
	return resolved, nil
}

func mergeIncludeExcludes(resolved *ResolvedMetricsViewSecurity, include, exclude []string, excludeAll bool) {
	if resolved.ExcludeAll {
		return
	}

	if excludeAll {
		resolved.Include = nil
		resolved.Exclude = nil
		resolved.ExcludeAll = true
		return
	}

	if len(resolved.Include) == 0 && len(resolved.Exclude) == 0 {
		resolved.Include = include
		resolved.Exclude = exclude
		return
	}

	if len(resolved.Include) == 0 && len(include) == 0 {
		resolved.Exclude = append(resolved.Exclude, exclude...)
		return
	}

	// Build a new include list that is the intersection of resolved.Include and include, minus any fields in exclude.
	var newInclude []string
	for _, f := range include {
		// Check it's also in resolved.Include
		found := false
		for _, f2 := range resolved.Include {
			if f == f2 {
				found = true
				break
			}
		}
		if !found {
			break
		}

		// Check it's not in resolved.Exclude
		found = false
		for _, f2 := range resolved.Exclude {
			if f == f2 {
				found = true
				break
			}
		}
		if found {
			break
		}

		// Can add to newInclude
		newInclude = append(newInclude, f)
	}

	// If newInclude is empty, exclude all fields
	if len(newInclude) == 0 {
		resolved.Include = nil
		resolved.Exclude = nil
		resolved.ExcludeAll = true
	} else {
		resolved.Include = newInclude
		resolved.Exclude = nil
	}
}
