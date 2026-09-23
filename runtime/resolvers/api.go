package resolvers

import (
	"context"
	"fmt"
	"maps"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolver("api", newAPI, analyzeAPI)
}

type apiProps struct {
	API  string         `mapstructure:"api"`
	Args map[string]any `mapstructure:"args"`
}

// newAPI creates a resolver that proxies to the resolver of an API.
func newAPI(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props, args, err := parseAPI(opts.Properties, opts.Args)
	if err != nil {
		return nil, err
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("api", props.API))
	}

	// Find the API
	api, err := opts.Runtime.APIForName(ctx, opts.InstanceID, props.API, opts.Claims)
	if err != nil {
		return nil, err
	}

	// We need to protect against infinite recursion where API A proxies to another API that proxies back to API A.
	// For convenience, we overload the args with a special key that we use to track the APIs we've visited.
	key := "__internal__apis_visited"
	visited, ok := args[key].([]string)
	if !ok {
		visited = []string{}
	}
	for _, v := range visited {
		if v == props.API {
			return nil, fmt.Errorf("infinite recursion detected: the API %q proxies to itself", v)
		}
	}
	visited = append(visited, props.API)
	args[key] = visited

	// Initialize the resolver of the API to proxy to
	initializer, ok := runtime.ResolverInitializers[api.Spec.Resolver]
	if !ok {
		return nil, fmt.Errorf("no resolver found of type %q", api.Spec.Resolver)
	}
	return initializer(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: api.Spec.ResolverProperties.AsMap(),
		Args:       args,
		Claims:     opts.Claims,
		ForExport:  opts.ForExport,
	})
}

func parseAPI(properties, arguments map[string]any) (*apiProps, map[string]any, error) {
	props := &apiProps{}
	if err := mapstructureutil.WeakDecode(properties, props); err != nil {
		return nil, nil, err
	}
	// Static API arguments take precedence; never modify the caller's arguments.
	args := maps.Clone(arguments)
	if args == nil {
		args = make(map[string]any)
	}
	maps.Copy(args, props.Args)
	return props, args, nil
}

type analysisAPIPathKey struct{}

type analysisAPIRef struct {
	instanceID string
	name       string
}

func analyzeAPI(ctx context.Context, rt *runtime.Runtime, opts *runtime.ResolverAnalysisOptions) (*runtime.ResolverAnalysis, error) {
	props, args, err := parseAPI(opts.Properties, opts.Args)
	if err != nil {
		return nil, err
	}
	ref := analysisAPIRef{instanceID: opts.InstanceID, name: props.API}
	path, _ := ctx.Value(analysisAPIPathKey{}).([]analysisAPIRef)
	if slices.Contains(path, ref) {
		return nil, fmt.Errorf("infinite recursion detected: the API %q proxies to itself", props.API)
	}
	ctx = context.WithValue(ctx, analysisAPIPathKey{}, append(slices.Clone(path), ref))

	// Analysis inspects definitions only; APIForName remains the authorized execution lookup.
	api, ok := runtime.BuiltinAPIs[props.API]
	if !ok {
		ctrl, err := rt.Controller(ctx, opts.InstanceID)
		if err != nil {
			return nil, err
		}
		res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindAPI, Name: props.API}, false)
		if err != nil {
			return nil, err
		}
		api = res.GetApi()
	}
	child := *opts
	child.Properties = api.Spec.ResolverProperties.AsMap()
	child.Args = args
	return rt.AnalyzeResolver(ctx, api.Spec.Resolver, &child)
}
