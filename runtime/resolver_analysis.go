package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ResolverAnalysisOptions configures analysis of a resolver without authorizing or executing it.
// UserAttributes are used only for templating; analysis never accepts security claims.
type ResolverAnalysisOptions struct {
	InstanceID     string
	Properties     map[string]any
	Args           map[string]any
	UserAttributes map[string]any
	ForExport      bool
}

// ResolverAnalysis describes the dependencies and security rules required by a resolver.
// It does not grant access; callers must still apply the resource's security policies.
type ResolverAnalysis struct {
	Refs                  []*runtimev1.ResourceName
	RequiredSecurityRules []*runtimev1.SecurityRule
}

// ResolverAnalyzer analyzes a resolver without constructing an executable resolver.
// Implementations must not mutate the options or resolve security.
// Only bounded timestamp lookups needed to interpret dynamic time expressions may query data.
type ResolverAnalyzer func(context.Context, *Runtime, *ResolverAnalysisOptions) (*ResolverAnalysis, error)

var resolverAnalyzers = make(map[string]ResolverAnalyzer)

var errAnalysisUnsupported = errors.New("security rule inference not implemented")

// AnalysisUnsupported is the analyzer for resolvers that cannot infer security rules.
// Such resolvers cannot be used for transitive access (e.g. in reports or alerts).
func AnalysisUnsupported(context.Context, *Runtime, *ResolverAnalysisOptions) (*ResolverAnalysis, error) {
	return nil, errAnalysisUnsupported
}

// AnalyzeResolver discovers dependencies and required security rules for internal transitive-access inference.
// Unsupported resolvers return an error; analysis never falls back to executing an initializer.
func (r *Runtime) AnalyzeResolver(ctx context.Context, name string, opts *ResolverAnalysisOptions) (*ResolverAnalysis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	analyzer, ok := resolverAnalyzers[name]
	if !ok {
		return nil, fmt.Errorf("no resolver found for name %q", name)
	}
	analysis, err := analyzer(ctx, r, opts)
	if err != nil {
		if errors.Is(err, errAnalysisUnsupported) {
			return nil, fmt.Errorf("%w for resolver %q", err, name)
		}
		return nil, err
	}
	slices.SortFunc(analysis.Refs, func(a, b *runtimev1.ResourceName) int {
		if a.Kind != b.Kind {
			return strings.Compare(a.Kind, b.Kind)
		}
		return strings.Compare(a.Name, b.Name)
	})
	analysis.Refs = slices.CompactFunc(analysis.Refs, func(a, b *runtimev1.ResourceName) bool {
		return a.Kind == b.Kind && a.Name == b.Name
	})
	return analysis, nil
}
