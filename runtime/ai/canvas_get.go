package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const GetCanvasName = "get_canvas"

type GetCanvas struct {
	Runtime *runtime.Runtime
}

var _ Tool[*GetCanvasArgs, *GetCanvasResult] = (*GetCanvas)(nil)

type GetCanvasArgs struct {
	Canvas string `json:"canvas" jsonschema:"Name of the canvas"`
}

type GetCanvasResult struct {
	Spec         map[string]any `json:"spec"`
	Components   map[string]any `json:"components"`
	MetricsViews map[string]any `json:"metrics_views"`
}

func (t *GetCanvas) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        GetCanvasName,
		Title:       "Get Canvas",
		Description: "Get the specification for a given canvas, including available components and metrics views",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Getting canvas definition...",
			"openai/toolInvocation/invoked":  "Found canvas definition",
		},
	}
}

func (t *GetCanvas) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadObjects), nil
}

func (t *GetCanvas) Handler(ctx context.Context, args *GetCanvasArgs) (*GetCanvasResult, error) {
	session := GetSession(ctx)

	resolvedCanvas, err := t.Runtime.ResolveCanvas(ctx, session.InstanceID(), args.Canvas, session.Claims())
	if err != nil {
		return nil, err
	}

	if resolvedCanvas == nil || resolvedCanvas.Canvas == nil {
		return nil, fmt.Errorf("canvas %q not found", args.Canvas)
	}
	canvasSpec := resolvedCanvas.Canvas.GetCanvas().State.ValidSpec
	if canvasSpec == nil {
		return nil, fmt.Errorf("canvas %q is not valid", args.Canvas)
	}

	specMap := protoToJSON(canvasSpec)

	var components, metricsViews map[string]any
	for name, res := range resolvedCanvas.ResolvedComponents {
		component := res.GetComponent()
		if component == nil {
			continue
		}
		components[name] = protoToJSON(component.State.ValidSpec)
	}

	for name, res := range resolvedCanvas.ReferencedMetricsViews {
		metricsView := res.GetMetricsView()
		if metricsView == nil {
			continue
		}
		metricsViews[name] = protoToJSON(metricsView.State.ValidSpec)
	}

	return &GetCanvasResult{
		Spec:         specMap,
		Components:   components,
		MetricsViews: metricsViews,
	}, nil
}

func protoToJSON(spec proto.Message) map[string]any {
	if spec == nil {
		return nil
	}

	specJson, err := protojson.Marshal(spec)
	if err != nil {
		return nil
	}

	var specMap map[string]any
	err = json.Unmarshal(specJson, &specMap)
	if err != nil {
		return nil
	}

	return specMap
}
