package rillv1

import (
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/structpb"

	_ "embed"
)

//go:embed data/component-template-v1.json
var componentTemplateSpec string

var componentTemplateSchema = jsonschema.MustCompileString("https://github.com/rilldata/rill/tree/main/runtime/compilers/rillv1/data/component-template-v1.json", componentTemplateSpec)

type ComponentYAML struct {
	commonYAML  `yaml:",inline"`          // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName string                    `yaml:"display_name"`
	Title       string                    `yaml:"title"` // Deprecated: use display_name
	Description string                    `yaml:"description"`
	Subtitle    string                    `yaml:"subtitle"` // Deprecated: use description
	Input       []*ComponentVariableYAML  `yaml:"input"`
	Output      *ComponentVariableYAML    `yaml:"output"`
	Show        string                    `yaml:"show"`
	Other       map[string]map[string]any `yaml:",inline" mapstructure:",remain"` // Generic renderer: can only have one key
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
	spec, refs, err := p.parseComponentYAML(tmp)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, refs...)

	// Track component
	r, err := p.insertResource(ResourceKindComponent, node.Name, node.Paths, node.Refs...)
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
func (p *Parser) parseComponentYAML(tmp *ComponentYAML) (*runtimev1.ComponentSpec, []ResourceName, error) {
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
	var rendererProps *structpb.Struct
	if len(tmp.Other) == 1 {
		n++
		var props map[string]any
		for renderer, props = range tmp.Other {
			break
		}

		// nolint // TODO: Activate validation later when in production
		if err := componentTemplateSchema.Validate(map[string]any{renderer: props}); err != nil {
			// return nil, nil, fmt.Errorf(`failed to validate renderer %q: %w`, renderer, err)
		}

		propsPB, err := structpb.NewStruct(props)
		if err != nil {
			return nil, nil, fmt.Errorf(`failed to convert property %q to struct: %w`, renderer, err)
		}

		rendererProps = propsPB
	} else {
		n += len(tmp.Other)
	}

	// We generally treat the renderer props as untyped, but since "metrics_view" is a very common field,
	// and adding it to refs generally makes for nicer error messages, we specifically search for and link it here.
	var refs []ResourceName
	for k, v := range rendererProps.Fields {
		if k == "metrics_view" {
			name := v.GetStringValue()
			if name != "" {
				refs = append(refs, ResourceName{Kind: ResourceKindMetricsView, Name: name})
			}
			break
		}
	}

	// Check there is exactly one renderer
	if n == 0 {
		return nil, nil, errors.New(`missing renderer configuration`)
	}
	if n > 1 {
		return nil, nil, errors.New(`multiple renderers are not allowed`)
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
		Input:              input,
		Output:             output,
		Show:               tmp.Show,
	}

	return spec, refs, nil
}

type ComponentVariableYAML struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Value any    `yaml:"value"`
}

func (y *ComponentVariableYAML) Proto() (*runtimev1.ComponentVariable, error) {
	if y == nil {
		return nil, fmt.Errorf("is empty")
	}
	val, err := structpb.NewValue(y.Value)
	if err != nil {
		panic(fmt.Errorf("invalid default value: %w", err))
	}
	return &runtimev1.ComponentVariable{
		Name:         y.Name,
		Type:         y.Type,
		DefaultValue: val,
	}, nil
}
