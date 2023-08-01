package runtime

import (
	"fmt"
	"hash/maphash"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"go.uber.org/zap"
)

type policyEngine struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	logger *zap.Logger
}

func newPolicyEngine(cacheSize int, logger *zap.Logger) *policyEngine {
	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		panic(err)
	}
	return &policyEngine{cache: cache, logger: logger}
}

var openPolicy = &ResolvedMetricsViewPolicy{
	HasAccess: true,
	Filter:    "",
	Include:   nil,
	Exclude:   nil,
}

type ResolvedMetricsViewPolicy struct {
	HasAccess bool
	Filter    string
	Include   []string
	Exclude   []string
}

func (p *policyEngine) resolveMetricsViewPolicy(attributes map[string]any, instanceID string, mv *runtimev1.MetricsView, lastUpdatedOn time.Time) (*ResolvedMetricsViewPolicy, error) {
	if mv.Policy == nil {
		return nil, nil
	}

	// if attributes is empty that means auth is disabled and also no user context is available
	// since we are controlling the attributes we can safely return the open policy
	if len(attributes) == 0 {
		return openPolicy, nil
	}

	key := fmt.Sprintf("%v:%v:%v", instanceID, mv.Name, lastUpdatedOn)
	for k, v := range attributes {
		key += fmt.Sprintf(":%v:%v", k, v)
	}
	var h maphash.Hash
	_, err := h.WriteString(key)
	if err != nil {
		return nil, err
	}
	cacheKey := h.Sum64()

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedMetricsViewPolicy), nil
	}

	resolved := &ResolvedMetricsViewPolicy{}
	templateData := &rillv1.TemplateData{Claims: attributes}

	if mv.Policy.HasAccess != "" {
		hasAccess, err := rillv1.ResolveTemplate(mv.Policy.HasAccess, *templateData)
		if err != nil {
			return nil, err
		}
		resolved.HasAccess, err = rillv1.EvaluateBoolExpression(hasAccess)
		if err != nil {
			return nil, err
		}
	}

	if mv.Policy.Filter != "" {
		filter, err := rillv1.ResolveTemplate(mv.Policy.Filter, *templateData)
		if err != nil {
			return nil, err
		}
		resolved.Filter = filter
	}

	for _, inc := range mv.Policy.Include {
		inc.Condition, err = rillv1.ResolveTemplate(inc.Condition, *templateData)
		if err != nil {
			return nil, err
		}
		incCond, err := rillv1.EvaluateBoolExpression(inc.Condition)
		if err != nil {
			return nil, err
		}
		if incCond {
			resolved.Include = append(resolved.Include, inc.Name)
		}
	}
	for _, exc := range mv.Policy.Exclude {
		exc.Condition, err = rillv1.ResolveTemplate(exc.Condition, *templateData)
		if err != nil {
			return nil, err
		}
		excCond, err := rillv1.EvaluateBoolExpression(exc.Condition)
		if err != nil {
			return nil, err
		}
		if excCond {
			resolved.Exclude = append(resolved.Exclude, exc.Name)
		}
	}

	p.cache.Add(cacheKey, resolved)
	return resolved, nil
}
