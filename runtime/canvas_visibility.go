package runtime

import (
	"context"
	"strings"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/parser"
	"google.golang.org/protobuf/proto"
)

// builtinUserAttributeDefaults are the values of the built-in user attributes when the claims don't carry them,
// such as for an embed token created from custom attributes.
// Visibility conditions can always reference these; referencing a missing custom attribute makes a condition false.
var builtinUserAttributeDefaults = map[string]any{
	"name":   "",
	"email":  "",
	"domain": "",
	"groups": []any{},
	"admin":  false,
	"embed":  false,
}

// VisibilityConditions evaluates the `if` conditions of canvases and components for one set of claims, and memoizes the results.
// A condition only depends on the claims, the environment and the variables, not on the canvas or component that declares it,
// so the results are shared by every canvas and component of the instance.
// A nil *VisibilityConditions meets every condition, which is used when security checks are skipped.
type VisibilityConditions struct {
	td      parser.TemplateData
	mu      sync.Mutex
	results map[string]bool
}

// VisibilityConditions returns the evaluator of `if` conditions for the claims.
// Unlike ResolveSecurity, it doesn't apply security rules,
// so it can compute the content of a canvas that transitive access grants cover without recursing into the security engine,
// and it also works for claims that are only granted a canvas (e.g. alert links).
func (r *Runtime) VisibilityConditions(ctx context.Context, instanceID string, claims *SecurityClaims) (*VisibilityConditions, error) {
	if claims.SkipChecks {
		return nil, nil
	}
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return r.securityEngine.visibilityConditions(instanceID, inst.Environment, inst.ResolveVariables(false), claims)
}

// newVisibilityConditions creates an evaluator of `if` conditions for the given template data,
// with defaults for the built-in user attributes the claims don't carry.
func newVisibilityConditions(td parser.TemplateData) *VisibilityConditions {
	user := make(map[string]any, len(td.User)+len(builtinUserAttributeDefaults))
	for k, v := range builtinUserAttributeDefaults {
		user[k] = v
	}
	for k, v := range td.User {
		user[k] = v
	}
	td.User = user
	return &VisibilityConditions{td: td, results: make(map[string]bool)}
}

// Met reports whether the claims meet the condition.
// An empty condition is always met.
// A condition that fails to evaluate, such as one referencing a user attribute the claims don't have, isn't met.
func (v *VisibilityConditions) Met(condition string) bool {
	if v == nil || condition == "" {
		return true
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if met, ok := v.results[condition]; ok {
		return met
	}
	met, _ := evaluateVisibilityCondition(condition, v.td)
	v.results[condition] = met
	return met
}

// Keep returns a canvas.Keep that keeps the rows, tabs and items whose `if` condition the claims meet.
func (v *VisibilityConditions) Keep() canvas.Keep {
	return canvas.Keep{
		Row:  func(row *runtimev1.CanvasRow) bool { return v.Met(row.ConditionExpression) },
		Tab:  func(tab *runtimev1.CanvasTab) bool { return v.Met(tab.ConditionExpression) },
		Item: func(item *runtimev1.CanvasItem) bool { return v.Met(item.ConditionExpression) },
	}
}

// ComponentMet reports whether the claims meet the `if` conditions that an inline canvas component inherits from its item and the item's ancestors.
// The component keeps those conditions wherever it's referenced, including by another canvas.
func (v *VisibilityConditions) ComponentMet(res *runtimev1.Resource) bool {
	spec := res.GetComponent().GetState().GetValidSpec()
	if spec == nil {
		spec = res.GetComponent().GetSpec()
	}
	return v.Met(spec.GetConditionExpression())
}

// applyCanvasSecurity removes the content of a canvas that `if` conditions hide from the claims, and removes the conditions themselves.
// The resource is cloned if anything changes.
func (r *Runtime) applyCanvasSecurity(res *runtimev1.Resource, security *ResolvedSecurity) *runtimev1.Resource {
	if security.visibility == nil || !canvasHasConditions(res) {
		return res
	}

	res = proto.Clone(res).(*runtimev1.Resource)
	keep := security.visibility.Keep()
	referenced := make(map[string]bool)
	for _, spec := range []*runtimev1.CanvasSpec{res.GetCanvas().GetSpec(), res.GetCanvas().GetState().GetValidSpec()} {
		if spec == nil {
			continue
		}
		spec.Rows, _ = canvas.PruneRows(spec.Rows, keep)
		canvas.StripConditions(spec.Rows)
		CollectCanvasComponentNames(spec.Rows, referenced)
	}

	// The refs also name the hidden components.
	refs := res.Meta.Refs[:0]
	for _, ref := range res.Meta.Refs {
		if ref.Kind == ResourceKindComponent && !referenced[ref.Name] {
			continue
		}
		refs = append(refs, ref)
	}
	res.Meta.Refs = refs

	return res
}

// stripComponentCondition clears the `if` conditions that an inline component inherits from its canvas item, so they aren't sent to clients.
// Only the security engine needs them, and it reads components from the controller.
// The resource is cloned if it has a condition.
func stripComponentCondition(res *runtimev1.Resource) *runtimev1.Resource {
	c := res.GetComponent()
	if c.GetSpec().GetConditionExpression() == "" && c.GetState().GetValidSpec().GetConditionExpression() == "" {
		return res
	}
	res = proto.Clone(res).(*runtimev1.Resource)
	if spec := res.GetComponent().GetSpec(); spec != nil {
		spec.ConditionExpression = ""
	}
	if spec := res.GetComponent().GetState().GetValidSpec(); spec != nil {
		spec.ConditionExpression = ""
	}
	return res
}

// canIncludeHiddenCanvasContent reports whether the claims may get a canvas including the content hidden from them, as the visual editor needs.
// That requires permission to edit the project, or owning the canvas if it's a personal (admin-managed) canvas.
func canIncludeHiddenCanvasContent(claims *SecurityClaims, res *runtimev1.Resource) bool {
	if claims.Can(EditRepo) {
		return true
	}
	spec := res.GetCanvas().GetState().GetValidSpec()
	if spec == nil {
		spec = res.GetCanvas().GetSpec()
	}
	if !isAdminManagedAnnotations(spec.GetAnnotations()) {
		return false
	}
	userID, _ := claims.UserAttributes["id"].(string)
	return userID != "" && userID == spec.GetAnnotations()["admin_owner_user_id"]
}

// canvasHasConditions reports whether the canvas's spec or valid spec has `if` conditions.
func canvasHasConditions(res *runtimev1.Resource) bool {
	return len(canvas.Conditions(res.GetCanvas().GetSpec().GetRows())) > 0 ||
		len(canvas.Conditions(res.GetCanvas().GetState().GetValidSpec().GetRows())) > 0
}

// evaluateVisibilityCondition evaluates an `if` condition, which renders to a boolean SQL expression like a security policy's `access`.
// Referencing a user attribute the template data doesn't have is an error, so callers should treat errors as a condition that isn't met.
// Conditions that render to a literal true or false are evaluated without DuckDB.
func evaluateVisibilityCondition(condition string, td parser.TemplateData) (bool, error) {
	resolved, err := parser.ResolveTemplate(condition, td, true)
	if err != nil {
		return false, err
	}
	switch strings.TrimSpace(resolved) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return parser.EvaluateBoolExpression(resolved)
}
