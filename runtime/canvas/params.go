package canvas

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/types/known/structpb"
)

// ParamScalarTypes are the param types that hold plain values.
// Scalar params may declare options and are injected as native Vega-Lite params at resolve time.
var ParamScalarTypes = []string{"string", "number", "boolean"}

// ParamFieldTypes are the param types that reference a field of a metrics view.
// They resolve their metrics view through a sibling param of type "metrics_view".
var ParamFieldTypes = []string{"measure", "dimension", "time_dimension"}

// EffectiveArgs merges a component's declared param defaults with the provided args.
// Provided args take precedence over declared param defaults,
// which in turn take precedence over legacy input variable defaults.
func EffectiveArgs(spec *runtimev1.ComponentSpec, args map[string]any) map[string]any {
	res := make(map[string]any, len(args)+len(spec.Params)+len(spec.Input))
	for _, v := range spec.Input {
		if v.DefaultValue != nil {
			res[v.Name] = v.DefaultValue.AsInterface()
		}
	}
	for _, p := range spec.Params {
		if p.Default != nil {
			res[p.Name] = p.Default.AsInterface()
		}
	}
	for k, v := range args {
		res[k] = v
	}
	return res
}

// IsMetricsViewParamName reports whether a param name follows the convention used
// to discover metrics view bindings before a referenced component has been loaded.
func IsMetricsViewParamName(name string) bool {
	return name == "metrics_view" || strings.HasSuffix(name, "_metrics_view")
}

// MetricsViewNamesFromBindings extracts statically bound metrics view names from a params map.
// Templated bindings are resolved later and cannot be registered as resource references here.
func MetricsViewNamesFromBindings(bindings map[string]any) []string {
	var names []string
	for key, value := range bindings {
		if !IsMetricsViewParamName(key) {
			continue
		}
		if name, ok := value.(string); ok && name != "" && !IsTemplated(name) {
			names = append(names, name)
		}
	}
	return names
}

// BoundMetricsViewNames returns the static metrics view names bound to a component's params of type "metrics_view",
// falling back to their defaults. Pass nil bound to get only the defaults.
// Templated values are skipped since they resolve at render time and cannot be referenced statically.
func BoundMetricsViewNames(params []*runtimev1.ComponentParam, bound map[string]any) []string {
	var names []string
	for _, p := range params {
		if p.Type != "metrics_view" {
			continue
		}
		v, _ := effectiveParamValue(p, bound)
		if name, ok := v.(string); ok && name != "" && !IsTemplated(name) {
			names = append(names, name)
		}
	}
	return names
}

// ValidateParamBindings validates values bound to a component's declared params.
// It checks unknown keys, missing required params, scalar types and option membership.
// When metricsViews is non-nil, it additionally checks that params of type "metrics_view" name a valid metrics view
// and that field-typed params reference fields of their bound metrics view.
// The provided metricsViews should contain every valid metrics view bound to a param (as determined by refs in the parser);
// a bound metrics view missing from the map is reported as invalid (don't look it up separately in the catalog).
// Pass nil metricsViews to skip the metrics view checks (e.g. at resolve time, where they already ran at reconcile time).
//
// Bound string values containing template placeholders (e.g. {{ .env.name }}) are not validated;
// they resolve at render time.
func ValidateParamBindings(params []*runtimev1.ComponentParam, bound map[string]any, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	declared := make(map[string]*runtimev1.ComponentParam, len(params))
	for _, p := range params {
		declared[p.Name] = p
	}

	for k := range bound {
		if declared[k] == nil {
			return fmt.Errorf("unknown param %q: it is not declared by the component", k)
		}
	}

	for _, p := range params {
		v, err := effectiveParamValue(p, bound)
		if err != nil {
			return err
		}
		if v == nil {
			// Optional and unbound with no default.
			continue
		}
		if IsTemplated(v) {
			continue
		}

		if err := ValidateParamValue(p.Type, v); err != nil {
			return fmt.Errorf("invalid value for param %q: %w", p.Name, err)
		}

		if len(p.Options) > 0 {
			val, err := structpb.NewValue(v)
			if err != nil {
				return fmt.Errorf("invalid value for param %q: %w", p.Name, err)
			}
			if !IsParamOption(val, p.Options) {
				return fmt.Errorf("value %v for param %q is not one of its options", v, p.Name)
			}
		}

		if metricsViews == nil {
			continue
		}

		switch p.Type {
		case "metrics_view":
			if metricsViews[v.(string)] == nil {
				return fmt.Errorf("metrics view %q bound to param %q is invalid or does not exist", v, p.Name)
			}
		case "measure", "dimension", "time_dimension":
			mvn, err := effectiveParamValue(declared[p.MetricsViewParam], bound)
			if err != nil || mvn == nil || IsTemplated(mvn) {
				// The metrics view param is itself invalid or unresolvable; it reports its own error.
				continue
			}
			mv := metricsViews[mvn.(string)]
			if mv == nil {
				continue
			}
			field := v.(string)
			switch p.Type {
			case "measure":
				if !metricsViewHasMeasure(mv, field) {
					return fmt.Errorf("value %q for param %q is not a measure in metrics view %q", field, p.Name, mvn)
				}
			case "dimension":
				if !metricsViewHasDimension(mv, field) {
					return fmt.Errorf("value %q for param %q is not a dimension in metrics view %q", field, p.Name, mvn)
				}
			case "time_dimension":
				if !metricsview.IsTimeDimension(mv, field) {
					return fmt.Errorf("value %q for param %q is not a time dimension in metrics view %q", field, p.Name, mvn)
				}
			}
		}
	}

	return nil
}

// InjectVegaParams injects the values of a component's scalar params as native Vega-Lite params
// into a Vega-Lite spec (a JSON object), making them usable in Vega expressions, predicates and extents.
// If the spec already declares a top-level param with the same name, only its "value" is overridden
// so author-supplied properties like "bind" and "expr" are preserved.
// Field-typed params are not injected since Vega-Lite cannot parameterize field names;
// they are substituted through templating instead.
// Returns the spec unchanged if there is nothing to inject.
func InjectVegaParams(vegaSpec string, params []*runtimev1.ComponentParam, args map[string]any) (string, error) {
	values := resolvedArgsOfType(params, args, ParamScalarTypes)
	if len(values) == 0 || strings.TrimSpace(vegaSpec) == "" {
		return vegaSpec, nil
	}

	var spec map[string]any
	if err := json.Unmarshal([]byte(vegaSpec), &spec); err != nil {
		return "", fmt.Errorf("vega_spec is not a valid JSON object: %w", err)
	}

	existing, _ := spec["params"].([]any)
	// Iterate params rather than values to inject in declaration order.
	for _, p := range params {
		name := p.Name
		if _, ok := values[name]; !ok {
			continue
		}
		found := false
		for _, e := range existing {
			if m, ok := e.(map[string]any); ok && m["name"] == name {
				m["value"] = values[name]
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, map[string]any{"name": name, "value": values[name]})
		}
	}
	spec["params"] = existing

	out, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// wholeStringParamRef matches a string consisting of exactly one param placeholder and nothing else.
var wholeStringParamRef = regexp.MustCompile(`^\s*\{\{\s*\.(?:params|args)\.(\w+)\s*\}\}\s*$`)

// CoerceScalarParams substitutes numeric and boolean param values into renderer properties
// with their original type, rather than as strings.
//
// Template resolution always yields strings, so a structured renderer property such as a Flint
// spec's "chartProperties: {innerRadius: {{ .params.inner_radius }}}" would otherwise reach the
// renderer as "60" instead of 60 and silently degrade. This only applies to values that consist of
// exactly one placeholder: partial interpolations like "LIMIT {{ .params.limit }}" are strings by
// nature and are left to normal templating.
//
// It returns a copy; the input is not modified.
func CoerceScalarParams(props map[string]any, params []*runtimev1.ComponentParam, args map[string]any) map[string]any {
	values := resolvedArgsOfType(params, args, []string{"number", "boolean"})
	if len(values) == 0 {
		return props
	}

	res, _ := coerceScalarParams(props, values).(map[string]any)
	return res
}

func coerceScalarParams(val any, values map[string]any) any {
	switch val := val.(type) {
	case string:
		m := wholeStringParamRef.FindStringSubmatch(val)
		if m == nil {
			return val
		}
		v, ok := values[m[1]]
		if !ok {
			return val
		}
		return v
	case map[string]any:
		res := make(map[string]any, len(val))
		for k, v := range val {
			res[k] = coerceScalarParams(v, values)
		}
		return res
	case []any:
		res := make([]any, len(val))
		for i, v := range val {
			res[i] = coerceScalarParams(v, values)
		}
		return res
	default:
		return val
	}
}

// resolvedArgsOfType returns the args bound to params of the given types, keyed by param name.
// Unbound and templated args are omitted.
func resolvedArgsOfType(params []*runtimev1.ComponentParam, args map[string]any, types []string) map[string]any {
	res := make(map[string]any)
	for _, p := range params {
		if !slices.Contains(types, p.Type) {
			continue
		}
		if v, ok := args[p.Name]; ok && v != nil && !IsTemplated(v) {
			res[p.Name] = v
		}
	}
	return res
}

// effectiveParamValue returns the value bound to a param, falling back to its default.
// It returns nil if the param is unbound and has no default, or an error if it is required and unbound.
func effectiveParamValue(p *runtimev1.ComponentParam, bound map[string]any) (any, error) {
	if p == nil {
		return nil, nil
	}
	if v, ok := bound[p.Name]; ok && v != nil {
		return v, nil
	}
	if p.Default != nil {
		return p.Default.AsInterface(), nil
	}
	if p.Required {
		return nil, fmt.Errorf("missing value for required param %q", p.Name)
	}
	return nil, nil
}

// ValidateParamValue checks that a param value (a binding, default or option) conforms to the param's declared type.
func ValidateParamValue(typ string, v any) error {
	switch typ {
	case "number":
		switch v.(type) {
		case int, int32, int64, uint, uint32, uint64, float32, float64:
			return nil
		}
		return fmt.Errorf("expected a number, got %v", v)
	case "boolean":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected a boolean, got %v", v)
		}
		return nil
	default:
		// "string" and the metrics view field types hold string values.
		if _, ok := v.(string); !ok {
			return fmt.Errorf("expected a string, got %v", v)
		}
		return nil
	}
}

// IsParamOption reports whether a scalar value is one of a param's options.
// Protobuf values hold every number as a float64, so all numeric types compare equal when their values match.
func IsParamOption(v *structpb.Value, options []*structpb.Value) bool {
	for _, opt := range options {
		if v.AsInterface() == opt.AsInterface() {
			return true
		}
	}
	return false
}

// IsTemplated returns true if the value is a string containing template placeholders.
func IsTemplated(v any) bool {
	s, ok := v.(string)
	return ok && strings.Contains(s, "{{")
}
