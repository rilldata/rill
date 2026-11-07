package ai

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

const ListCanvasesName = "list_canvases"

type ListCanvases struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListCanvasesArgs, *ListCanvasesResult] = (*ListCanvases)(nil)

type ListCanvasesArgs struct{}

type ListCanvasesResult struct {
	Canvases []map[string]any `json:"canvases"`
}

func (t *ListCanvases) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListCanvasesName,
		Title:       "List Canvases",
		Description: "List the canvas dashboards in the current project that the user can see, including the user's own personal dashboards, which are marked with personal: true",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing dashboards...",
			"openai/toolInvocation/invoked":  "Listed dashboards",
		},
	}
}

func (t *ListCanvases) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadObjects), nil
}

func (t *ListCanvases) Handler(ctx context.Context, args *ListCanvasesArgs) (*ListCanvasesResult, error) {
	session := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, session.InstanceID())
	if err != nil {
		return nil, err
	}

	rs, err := ctrl.List(ctx, runtime.ResourceKindCanvas, "", false)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		return strings.Compare(a.Meta.Name.Name, b.Meta.Name.Name)
	})

	canvases := []map[string]any{}
	for _, r := range rs {
		r, access, err := t.Runtime.ApplySecurityPolicy(ctx, session.InstanceID(), session.Claims(), r)
		if err != nil {
			return nil, err
		}
		if !access {
			continue
		}

		// List a personal canvas with errors too, so the user can ask to fix it.
		spec := r.GetCanvas().State.ValidSpec
		if spec == nil {
			spec = r.GetCanvas().Spec
		}

		// Personal canvases are stored by the admin service and owned by a user.
		// Only list the user's own: project admins can open other users' personal canvases, but they are not theirs to browse.
		personal, _ := strconv.ParseBool(spec.Annotations["admin_managed"])
		if personal && (session.Claims().UserID == "" || spec.Annotations["admin_owner_user_id"] != session.Claims().UserID) {
			continue
		}

		canvases = append(canvases, map[string]any{
			"name":         r.Meta.Name.Name,
			"display_name": spec.DisplayName,
			"personal":     personal,
		})
	}

	return &ListCanvasesResult{
		Canvases: canvases,
	}, nil
}
