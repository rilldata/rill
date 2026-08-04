package parser

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

type ComponentYAML struct {
	commonYAML  `yaml:",inline"`          // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName string                    `yaml:"display_name"`
	Title       string                    `yaml:"title"` // Deprecated: use display_name
	Description string                    `yaml:"description"`
	Subtitle    string                    `yaml:"subtitle"` // Deprecated: use description
	Params      []*ComponentParamYAML     `yaml:"params"`
	Input       []*ComponentVariableYAML  `yaml:"input"`
	Output      *ComponentVariableYAML    `yaml:"output"`
	Other       map[string]map[string]any `yaml:",inline" mapstructure:",remain"` // Generic renderer: can only have one key
}

// ComponentParamYAML declares a typed parameter of a component.
// Canvases bind values to params when referencing the component;
// bound values are available in the renderer properties' templating as {{ .params.<name> }}.
type ComponentParamYAML struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default"`
	MetricsView string `yaml:"metrics_view"` // For field-typed params: back-reference to a sibling param of type "metrics_view"
	Options     []any  `yaml:"options"`      // For scalar params: allowed values
}

// componentParamScalarTypes are the param types that hold plain values.
// Scalar params may declare options and are injected as native Vega-Lite params at resolve time.
var componentParamScalarTypes = []string{"string", "number", "boolean"}

// componentParamFieldTypes are the param types that reference a field of a metrics view.
// They resolve their metrics view through a sibling param of type "metrics_view".
var componentParamFieldTypes = []string{"measure", "dimension", "time_dimension"}

// componentParamNameRegex restricts param names to valid identifiers,
// since they are referenced as {{ .params.<name> }} in Go templates and as signal names in Vega expressions.
var componentParamNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// componentParamReservedNames are names that cannot be used for params:
// "title" and "description" collide with common component properties in visual editors,
// and "datum", "item", "event", "parent" are reserved in Vega expressions
// (scalar params are injected as native Vega-Lite params at resolve time).
var componentParamReservedNames = []string{"title", "description", "datum", "item", "event", "parent"}

type ComponentVariableYAML struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Value any    `yaml:"value"`
}

func (y *ComponentVariableYAML) Proto() (*runtimev1.ComponentVariable, error) {
	if y == nil {
		return nil, fmt.Errorf("is empty")
	}
	val, err := pbutil.ToValue(y.Value, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid default value: %w", err)
	}
	return &runtimev1.ComponentVariable{
		Name:         y.Name,
		Type:         y.Type,
		DefaultValue: val,
	}, nil
}

func (p *Parser) parseComponent(node *Node) error {
	// Parse YAML
	tmp := &ComponentYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("components cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("components cannot have a connector")
	}

	// Parse into a ComponentSpec
	spec, refs, err := p.parseComponentYAML(tmp, false)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, refs...)

	// Track component
	r, err := p.insertResource(ResourceKindComponent, node.Name, node.Paths, node.Tags, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ComponentSpec = spec
	if r.ComponentSpec.DisplayName == "" {
		r.ComponentSpec.DisplayName = ToDisplayName(node.Name)
	}

	return nil
}

// parseComponentYAML parses and validates a ComponentYAML.
// It is separated from parseComponent to allow inline creation of components from a canvas YAML file.
// The inline flag selects the custom chart spec format: inline canvas charts author Vega-Lite under
// "vega_spec", while standalone component files author Flint under "spec".
func (p *Parser) parseComponentYAML(tmp *ComponentYAML, inline bool) (*runtimev1.ComponentSpec, []ResourceName, error) {
	// Display name backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}

	// Description backwards compatibility
	if tmp.Subtitle != "" && tmp.Description == "" {
		tmp.Description = tmp.Subtitle
	}

	// Discover and validate the renderer
	n := 0
	var renderer string
	var props map[string]any
	var rendererProps *structpb.Struct
	if len(tmp.Other) == 1 {
		n++
		for renderer, props = range tmp.Other {
			break
		}

		propsPB, err := structpb.NewStruct(props)
		if err != nil {
			return nil, nil, fmt.Errorf(`failed to convert property %q to struct: %w`, renderer, err)
		}

		rendererProps = propsPB
	} else {
		n += len(tmp.Other)
	}

	// Check there is exactly one renderer
	if n == 0 {
		return nil, nil, errors.New(`missing renderer configuration`)
	}
	if n > 1 {
		return nil, nil, errors.New(`multiple renderers are not allowed`)
	}

	// Enforce the custom chart spec format for this context. Presence is not required:
	// the visual editor persists drafts with empty renderer properties.
	if renderer == "custom_chart" {
		if inline {
			if _, ok := props["spec"]; ok {
				return nil, nil, errors.New(`renderer property "spec" is not supported in an inline canvas chart: use "vega_spec", or define a component file to author a chart spec`)
			}
		} else {
			if _, ok := props["vega_spec"]; ok {
				return nil, nil, errors.New(`renderer property "vega_spec" is not supported in a component file: use "spec" with a chart spec`)
			}
			if spec, ok := props["spec"]; ok {
				if _, ok := spec.(map[string]any); !ok {
					return nil, nil, errors.New(`renderer property "spec" must be a mapping`)
				}
			}
		}
	}

	// We generally treat the renderer props as untyped, but since "metrics_view" is a very common field,
	// and adding it to refs generally makes for nicer error messages, we specifically search for and link it here.
	var refs []ResourceName
	if rendererProps != nil {
		for k, v := range rendererProps.Fields {
			if k == "metrics_view" {
				name := v.GetStringValue()
				// Skip templated values (e.g. {{ .params.metrics_view }}): refs must be static resource names.
				if name != "" && !strings.Contains(name, "{{") {
					refs = append(refs, ResourceName{Kind: ResourceKindMetricsView, Name: name})
				}
				break
			}
		}
	}

	// Parse declared params
	params, err := parseComponentParams(tmp.Params)
	if err != nil {
		return nil, nil, err
	}

	// Link metrics views set as defaults of metrics_view params:
	// a component rendered with its defaults depends on the default metrics view
	// even when no canvas binds the param, so it needs a ref for DAG ordering and invalidation.
	for _, param := range params {
		if param.Type != "metrics_view" || param.Default == nil {
			continue
		}
		if name, ok := param.Default.AsInterface().(string); ok && name != "" && !strings.Contains(name, "{{") {
			refs = append(refs, ResourceName{Kind: ResourceKindMetricsView, Name: name})
		}
	}

	// When params are declared, require that all template references to .params (or its .args alias)
	// in the renderer properties refer to declared params. This catches typos at parse time.
	if len(params) > 0 {
		declared := make(map[string]bool, len(params))
		for _, param := range params {
			declared[param.Name] = true
		}
		vars := make(map[string]string)
		if err := AnalyzeTemplateRecursively(props, vars); err != nil {
			return nil, nil, fmt.Errorf("failed to analyze templating in renderer properties: %w", err)
		}
		for v := range vars {
			rest, ok := strings.CutPrefix(v, "params.")
			if !ok {
				rest, ok = strings.CutPrefix(v, "args.")
			}
			if !ok {
				continue
			}
			name, _, _ := strings.Cut(rest, ".")
			if !declared[name] {
				return nil, nil, fmt.Errorf("renderer properties reference undeclared param %q", name)
			}
		}
	}

	// Parse input variables
	var input []*runtimev1.ComponentVariable
	if len(tmp.Input) > 0 {
		input = make([]*runtimev1.ComponentVariable, len(tmp.Input))
	}
	for i, v := range tmp.Input {
		var err error
		input[i], err = v.Proto()
		if err != nil {
			return nil, nil, fmt.Errorf("invalid input variable at index %d: %w", i, err)
		}
	}

	// Parse the output variable
	var output *runtimev1.ComponentVariable
	if tmp.Output != nil {
		var err error
		output, err = tmp.Output.Proto()
		if err != nil {
			return nil, nil, fmt.Errorf("invalid output variable: %w", err)
		}
	}

	// Create the component spec
	spec := &runtimev1.ComponentSpec{
		DisplayName:        tmp.DisplayName,
		Description:        tmp.Description,
		Renderer:           renderer,
		RendererProperties: rendererProps,
		Params:             params,
		Input:              input,
		Output:             output,
	}

	return spec, refs, nil
}

// parseComponentParams validates a component's declared params and converts them to protos.
func parseComponentParams(params []*ComponentParamYAML) ([]*runtimev1.ComponentParam, error) {
	if len(params) == 0 {
		return nil, nil
	}

	// First pass: validate names and collect params of type "metrics_view" for back-reference resolution.
	var mvParams []string
	seen := make(map[string]bool, len(params))
	for i, param := range params {
		if param == nil {
			return nil, fmt.Errorf("param at index %d is empty", i)
		}
		if !componentParamNameRegex.MatchString(param.Name) {
			return nil, fmt.Errorf("invalid param name %q: must be a valid identifier", param.Name)
		}
		if slices.Contains(componentParamReservedNames, param.Name) {
			return nil, fmt.Errorf("param name %q is reserved", param.Name)
		}
		if seen[param.Name] {
			return nil, fmt.Errorf("duplicate param name %q", param.Name)
		}
		seen[param.Name] = true

		if param.Type == "metrics_view" {
			// The naming convention enables the canvas parser to extract metrics view refs from
			// param bindings by key pattern without access to the component's declarations.
			if param.Name != "metrics_view" && !strings.HasSuffix(param.Name, "_metrics_view") {
				return nil, fmt.Errorf(`param %q of type "metrics_view" must be named "metrics_view" or end with "_metrics_view"`, param.Name)
			}
			mvParams = append(mvParams, param.Name)
		}
	}

	// Second pass: validate types, defaults, options and back-references, and convert to protos.
	res := make([]*runtimev1.ComponentParam, len(params))
	for i, param := range params {
		isScalar := slices.Contains(componentParamScalarTypes, param.Type)
		isField := slices.Contains(componentParamFieldTypes, param.Type)
		if !isScalar && !isField && param.Type != "metrics_view" {
			return nil, fmt.Errorf("param %q has invalid type %q (options: %s)", param.Name, param.Type, strings.Join(slices.Concat(componentParamScalarTypes, []string{"metrics_view"}, componentParamFieldTypes), ", "))
		}

		if param.Required && param.Default != nil {
			return nil, fmt.Errorf("param %q cannot both be required and have a default", param.Name)
		}
		if param.Default != nil {
			if err := validateComponentParamValue(param.Type, param.Default); err != nil {
				return nil, fmt.Errorf("invalid default for param %q: %w", param.Name, err)
			}
		}

		if len(param.Options) > 0 {
			if !isScalar {
				return nil, fmt.Errorf("param %q cannot have options: options are only supported for scalar types (%s)", param.Name, strings.Join(componentParamScalarTypes, ", "))
			}
			defaultInOptions := param.Default == nil
			for j, opt := range param.Options {
				if err := validateComponentParamValue(param.Type, opt); err != nil {
					return nil, fmt.Errorf("invalid option at index %d for param %q: %w", j, param.Name, err)
				}
				if param.Default != nil && scalarsEqual(param.Default, opt) {
					defaultInOptions = true
				}
			}
			if !defaultInOptions {
				return nil, fmt.Errorf("default for param %q is not one of its options", param.Name)
			}
		}

		mvParam := param.MetricsView
		if isField {
			if mvParam == "" {
				switch len(mvParams) {
				case 1:
					mvParam = mvParams[0]
				case 0:
					return nil, fmt.Errorf(`param %q of type %q requires a param of type "metrics_view" to be declared`, param.Name, param.Type)
				default:
					return nil, fmt.Errorf(`param %q of type %q must set "metrics_view" to disambiguate between multiple metrics_view params`, param.Name, param.Type)
				}
			} else if !slices.Contains(mvParams, mvParam) {
				return nil, fmt.Errorf(`param %q references %q, which is not a declared param of type "metrics_view"`, param.Name, mvParam)
			}
		} else if mvParam != "" {
			return nil, fmt.Errorf(`param %q of type %q cannot set "metrics_view"`, param.Name, param.Type)
		}

		var defaultVal *structpb.Value
		if param.Default != nil {
			var err error
			defaultVal, err = structpb.NewValue(param.Default)
			if err != nil {
				return nil, fmt.Errorf("invalid default for param %q: %w", param.Name, err)
			}
		}
		var options []*structpb.Value
		for j, opt := range param.Options {
			val, err := structpb.NewValue(opt)
			if err != nil {
				return nil, fmt.Errorf("invalid option at index %d for param %q: %w", j, param.Name, err)
			}
			options = append(options, val)
		}

		res[i] = &runtimev1.ComponentParam{
			Name:             param.Name,
			Type:             param.Type,
			Description:      param.Description,
			Required:         param.Required,
			Default:          defaultVal,
			MetricsViewParam: mvParam,
			Options:          options,
		}
	}

	return res, nil
}

// validateComponentParamValue checks that a param's default or option value conforms to its declared type.
func validateComponentParamValue(typ string, v any) error {
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
		// "string" and all metrics view field types hold string values.
		if _, ok := v.(string); !ok {
			return fmt.Errorf("expected a string, got %v", v)
		}
		return nil
	}
}

// scalarsEqual compares two scalar values, treating all numeric types as equal when their values match.
func scalarsEqual(a, b any) bool {
	af, aok := asFloat(a)
	bf, bok := asFloat(b)
	if aok && bok {
		return af == bf
	}
	return a == b
}

func asFloat(v any) (float64, bool) {
	switch v := v.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	}
	return 0, false
}
