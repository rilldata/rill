package runtime

import (
	"context"
	"strings"

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

// EvaluateCanvasConditions evaluates the `if` conditions of a canvas for the claims, keyed by condition.
// Unlike ResolveSecurity, it doesn't apply the canvas's security rules,
// so it can compute the content of a canvas that transitive access grants cover without recursing into the security engine.
// It returns nil when security checks are skipped, which keeps all content (see CanvasConditionsKeep).
func (r *Runtime) EvaluateCanvasConditions(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) (map[string]bool, error) {
	if claims.SkipChecks {
		return nil, nil
	}
	td, err := r.claimsTemplateData(ctx, instanceID, claims, res)
	if err != nil {
		return nil, err
	}
	return evaluateCanvasConditions(res, td), nil
}

// ComponentConditionMet reports whether the claims meet the `if` conditions that an inline canvas component inherits from its item.
// The component keeps those conditions wherever it's referenced, including by another canvas.
// Like EvaluateCanvasConditions, it doesn't apply security rules, so it also works for claims that are only granted a canvas (e.g. alert links).
func (r *Runtime) ComponentConditionMet(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) (bool, error) {
	if claims.SkipChecks {
		return true, nil
	}
	spec := res.GetComponent().GetState().GetValidSpec()
	if spec == nil {
		spec = res.GetComponent().GetSpec()
	}
	if spec.GetConditionExpression() == "" {
		return true, nil
	}
	td, err := r.claimsTemplateData(ctx, instanceID, claims, res)
	if err != nil {
		return false, err
	}
	return componentConditionMet(res, td), nil
}

// CanvasConditionsKeep returns a canvas.Keep that keeps the rows, tabs and items whose `if` condition is met according to results.
// Elements without a condition are always kept, and so is everything when results is nil.
// A condition missing from a non-nil results map is treated as not met.
func CanvasConditionsKeep(results map[string]bool) canvas.Keep {
	met := func(condition string) bool {
		if condition == "" || results == nil {
			return true
		}
		return results[condition]
	}
	return canvas.Keep{
		Row:  func(row *runtimev1.CanvasRow) bool { return met(row.ConditionExpression) },
		Tab:  func(tab *runtimev1.CanvasTab) bool { return met(tab.ConditionExpression) },
		Item: func(item *runtimev1.CanvasItem) bool { return met(item.ConditionExpression) },
	}
}

// applyCanvasSecurity removes the content of a canvas that `if` conditions hide from the claims, and removes the conditions themselves.
// The resource is cloned if anything changes.
func (r *Runtime) applyCanvasSecurity(res *runtimev1.Resource, security *ResolvedSecurity) *runtimev1.Resource {
	if len(security.canvasConditions) == 0 {
		return res
	}

	res = proto.Clone(res).(*runtimev1.Resource)
	keep := CanvasConditionsKeep(security.canvasConditions)
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

// claimsTemplateData returns the template data for evaluating the `if` conditions of a canvas or component for the claims.
func (r *Runtime) claimsTemplateData(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) (parser.TemplateData, error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return parser.TemplateData{}, err
	}
	return parser.TemplateData{
		Environment: inst.Environment,
		User:        claims.UserAttributes,
		Variables:   inst.ResolveVariables(false),
		Self:        parser.TemplateResource{Meta: res.Meta},
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
	}, nil
}

// evaluateCanvasConditions evaluates the distinct `if` conditions of the canvas's spec and valid spec.
func evaluateCanvasConditions(res *runtimev1.Resource, td parser.TemplateData) map[string]bool {
	td = visibilityTemplateData(td)
	results := make(map[string]bool)
	for _, spec := range []*runtimev1.CanvasSpec{res.GetCanvas().GetSpec(), res.GetCanvas().GetState().GetValidSpec()} {
		for _, condition := range canvas.Conditions(spec.GetRows()) {
			if _, ok := results[condition]; ok {
				continue
			}
			results[condition], _ = evaluateVisibilityCondition(condition, td)
		}
	}
	return results
}

// componentConditionMet reports whether the claims meet the `if` conditions that an inline canvas component inherits from its item and the item's ancestors.
func componentConditionMet(res *runtimev1.Resource, td parser.TemplateData) bool {
	spec := res.GetComponent().GetState().GetValidSpec()
	if spec == nil {
		spec = res.GetComponent().GetSpec()
	}
	condition := spec.GetConditionExpression()
	if condition == "" {
		return true
	}
	met, _ := evaluateVisibilityCondition(condition, visibilityTemplateData(td))
	return met
}

// visibilityTemplateData returns the template data for evaluating visibility conditions,
// with defaults for the built-in user attributes the claims don't carry.
func visibilityTemplateData(td parser.TemplateData) parser.TemplateData {
	user := make(map[string]any, len(td.User)+len(builtinUserAttributeDefaults))
	for k, v := range builtinUserAttributeDefaults {
		user[k] = v
	}
	for k, v := range td.User {
		user[k] = v
	}
	td.User = user
	return td
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
