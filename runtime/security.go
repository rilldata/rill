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
	Access:    true,
	RowFilter: "",
	Include:   nil,
	Exclude:   nil,
}

type ResolvedMetricsViewSecurity struct {
	Access     bool
	RowFilter  string
	Include    []string
	Exclude    []string
	ExcludeAll bool
}

func computeCacheKey(instanceID string, mv *runtimev1.MetricsViewSpec, lastUpdatedOn time.Time, attributes map[string]any) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(instanceID))
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(mv.Table))
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

func (p *securityEngine) resolveMetricsViewSecurity(instanceID, environment string, mv *runtimev1.MetricsViewSpec, lastUpdatedOn time.Time, attributes map[string]any) (*ResolvedMetricsViewSecurity, error) {
	if mv.Security == nil {
		return nil, nil
	}

	// if attributes is empty that means auth is disabled and also no user context is available
	// since we are controlling the attributes we can safely return the open policy
	if len(attributes) == 0 {
		return openAccess, nil
	}

	cacheKey, err := computeCacheKey(instanceID, mv, lastUpdatedOn, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to compute cache key: %w", err)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	cached, ok := p.cache.Get(cacheKey)
	if ok {
		return cached.(*ResolvedMetricsViewSecurity), nil
	}

	resolved := &ResolvedMetricsViewSecurity{}
	templateData := rillv1.TemplateData{
		Environment: environment,
		User:        attributes,
	}

	if mv.Security.Access != "" {
		access, err := rillv1.ResolveTemplate(mv.Security.Access, templateData)
		if err != nil {
			return nil, err
		}
		resolved.Access, err = rillv1.EvaluateBoolExpression(access)
		if err != nil {
			return nil, err
		}
	}

	if mv.Security.RowFilter != "" {
		filter, err := rillv1.ResolveTemplate(mv.Security.RowFilter, templateData)
		if err != nil {
			return nil, err
		}
		resolved.RowFilter = filter
	}

	seen := map[string]bool{}

	for _, inc := range mv.Security.Include {
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
				resolved.Include = append(resolved.Include, name)
			}
		}
	}

	// this is to handle the case where include filter was present but none of them evaluted to true
	if len(mv.Security.Include) > 0 && len(resolved.Include) == 0 {
		resolved.ExcludeAll = true
	}

	for _, exc := range mv.Security.Exclude {
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
				resolved.Exclude = append(resolved.Exclude, name)
			}
		}
	}

	p.cache.Add(cacheKey, resolved)
	return resolved, nil
}
