package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
)

// translatableKinds lists the resource kinds that can be translated and, for each, the base labels that may be translated.
// A name in a translation file is resolved against these kinds in order, so the order must match inferAmbiguousRefs in the parser.
// This lives in the reconciler because the parser can't know the kind of a name in a translation file.
var translatableKinds = []struct {
	Kind       string
	BaseLabels []string
}{
	{runtime.ResourceKindMetricsView, []string{parser.TranslationLabelDisplayName, parser.TranslationLabelDescription}},
	{runtime.ResourceKindExplore, []string{parser.TranslationLabelDisplayName, parser.TranslationLabelDescription}},
}

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindTranslation, newTranslationReconciler)
}

type TranslationReconciler struct {
	C *runtime.Controller
}

func newTranslationReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &TranslationReconciler{C: c}, nil
}

func (r *TranslationReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *TranslationReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetTranslation()
	b := to.GetTranslation()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *TranslationReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetTranslation()
	b := to.GetTranslation()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *TranslationReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetTranslation().State = &runtimev1.TranslationState{}
	return nil
}

// Reconcile validates that every translation in the spec targets a resource, dimension or measure that actually exists.
// It doesn't produce any state: the translations are served straight from the spec by runtime.ApplyTranslations.
//
// Note there's deliberately no spec hash check. Validity depends on the referenced resources, not just on the spec,
// so an early return would mask keys that became invalid when a dimension was renamed.
func (r *TranslationReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	t := self.GetTranslation()
	if t == nil {
		return runtime.ReconcileResult{Err: errors.New("not a translation")}
	}

	// Exit early for deletion
	if self.Meta.DeletedOn != nil {
		return runtime.ReconcileResult{}
	}

	// Validate against idle, error-free specs instead of half-reconciled ones.
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Report all the problems at once instead of one per reconcile.
	var errs []string
	for _, name := range slices.Sorted(maps.Keys(t.Spec.GetResources())) {
		err := r.validateResourceTranslation(ctx, name, t.Spec.Resources[name])
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return runtime.ReconcileResult{Err: errors.New(strings.Join(errs, "; "))}
	}

	return runtime.ReconcileResult{}
}

func (r *TranslationReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetTranslation() == nil {
		return nil, fmt.Errorf("not a translation resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// validateResourceTranslation checks that the translations for one resource name can actually be applied.
func (r *TranslationReconciler) validateResourceTranslation(ctx context.Context, name string, t *runtimev1.TranslationSpec_ResourceTranslation) error {
	// Find the translatable resource that the name refers to.
	var res *runtimev1.Resource
	var baseLabels []string
	for _, k := range translatableKinds {
		tmp, err := r.C.Get(ctx, &runtimev1.ResourceName{Kind: k.Kind, Name: name}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return err
		}
		res = tmp
		baseLabels = k.BaseLabels
		break
	}
	if res == nil {
		return fmt.Errorf("%q does not match a translatable resource", name)
	}

	// Catalog lookups are case-insensitive, but translations are applied on an exact name match.
	if res.Meta.Name.Name != name {
		return fmt.Errorf("%q does not match the name of %s/%s exactly", name, runtime.PrettifyResourceKind(res.Meta.Name.Kind), res.Meta.Name.Name)
	}

	// Check the base labels.
	for _, key := range slices.Sorted(maps.Keys(t.GetBaseTranslation().GetLabels())) {
		if !slices.Contains(baseLabels, key) {
			return fmt.Errorf("%q: %s does not have a translatable field named %q", name, runtime.PrettifyResourceKind(res.Meta.Name.Kind), key)
		}
	}

	if len(t.SubTranslations) == 0 {
		return nil
	}

	// Only metrics views have sub-entities to translate.
	mv := res.GetMetricsView()
	if mv == nil {
		return fmt.Errorf("%q: dimensions and measures can't be translated on %s", name, runtime.PrettifyResourceKind(res.Meta.Name.Kind))
	}

	for _, key := range slices.Sorted(maps.Keys(t.SubTranslations)) {
		switch t.SubTranslations[key].GetType() {
		case runtimev1.TranslationSpec_LABELS_TYPE_DIMENSION:
			if !slices.ContainsFunc(mv.State.GetValidSpec().GetDimensions(), func(d *runtimev1.MetricsViewSpec_Dimension) bool { return d.Name == key }) {
				return fmt.Errorf("%q: %q is not a dimension of the metrics view", name, key)
			}
		case runtimev1.TranslationSpec_LABELS_TYPE_MEASURE:
			if !slices.ContainsFunc(mv.State.GetValidSpec().GetMeasures(), func(m *runtimev1.MetricsViewSpec_Measure) bool { return m.Name == key }) {
				return fmt.Errorf("%q: %q is not a measure of the metrics view", name, key)
			}
		default:
			// The parser always sets a type, so this can only happen for a hand-written spec.
			return fmt.Errorf("%q: %q does not specify whether it translates a dimension or a measure", name, key)
		}
	}

	return nil
}
