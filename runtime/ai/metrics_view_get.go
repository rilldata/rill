package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

type GetMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[*GetMetricsViewArgs, *GetMetricsViewResult] = (*GetMetricsView)(nil)

type GetMetricsViewArgs struct {
	MetricsView string `json:"metrics_view" jsonschema:"Name of the metrics view"`
}

type GetMetricsViewResult struct {
	Spec map[string]any `json:"spec"`
}

func (t *GetMetricsView) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_metrics_view",
		Title:       "Get Metrics View",
		Description: "Get the specification for a given metrics view, including available measures and dimensions",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Getting metrics definition…",
			"openai/toolInvocation/invoked":  "Found metrics definition",
		},
	}
}

func (t *GetMetricsView) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadObjects)
}

func (t *GetMetricsView) Handler(ctx context.Context, args *GetMetricsViewArgs) (*GetMetricsViewResult, error) {
	session := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, session.InstanceID())
	if err != nil {
		return nil, err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: args.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	r, access, err := t.Runtime.ApplySecurityPolicy(ctx, session.InstanceID(), session.Claims(), r)
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, fmt.Errorf("resource not found")
	}

	if r.GetMetricsView().State.ValidSpec == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", args.MetricsView)
	}

	specJSON, err := protojson.Marshal(r.GetMetricsView().State.ValidSpec)
	if err != nil {
		return nil, err
	}
	var specMap map[string]any
	err = json.Unmarshal(specJSON, &specMap)
	if err != nil {
		return nil, err
	}

	return &GetMetricsViewResult{
		Spec: specMap,
	}, nil
}
