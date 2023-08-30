package runtime

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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

func computeCacheKey(instanceID string, mv *runtimev1.MetricsView, lastUpdatedOn time.Time, attributes map[string]any) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(instanceID))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(mv.Name))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(lastUpdatedOn.String()))
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
	return hex.EncodeToString(hash.Sum(nil)), nil
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

	cacheKey, err := computeCacheKey(instanceID, mv, lastUpdatedOn, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedMetricsViewPolicy), nil
	}

	resolved := &ResolvedMetricsViewPolicy{}
	templateData := rillv1.TemplateData{User: attributes}

	if mv.Policy.HasAccess != "" {
		hasAccess, err := rillv1.ResolveTemplate(mv.Policy.HasAccess, templateData)
		if err != nil {
			return nil, err
		}
		resolved.HasAccess, err = rillv1.EvaluateBoolExpression(hasAccess)
		if err != nil {
			return nil, err
		}
	}

	if mv.Policy.Filter != "" {
		filter, err := rillv1.ResolveTemplate(mv.Policy.Filter, templateData)
		if err != nil {
			return nil, err
		}
		resolved.Filter = filter
	}

	for _, inc := range mv.Policy.Include {
		cond, err := rillv1.ResolveTemplate(inc.Condition, templateData)
		if err != nil {
			return nil, err
		}
		incCond, err := rillv1.EvaluateBoolExpression(cond)
		if err != nil {
			return nil, err
		}
		if incCond {
			resolved.Include = append(resolved.Include, inc.Name)
		}
	}
	for _, exc := range mv.Policy.Exclude {
		cond, err := rillv1.ResolveTemplate(exc.Condition, templateData)
		if err != nil {
			return nil, err
		}
		excCond, err := rillv1.EvaluateBoolExpression(cond)
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
