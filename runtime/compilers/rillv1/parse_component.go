package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/structpb"

	_ "embed"
)

//go:embed data/vega-lite-v5.json
var vegaLiteSpec string

var vegaLiteSchema = jsonschema.MustCompileString("https://vega.github.io/schema/vega-lite/v5.json", vegaLiteSpec)

//go:embed data/component-template-v1.json
var componentTemplateSpec string

var componentTemplateSchema = jsonschema.MustCompileString("https://github.com/rilldata/rill/runtime/compilers/rillv1/data/component-template-v1.json", componentTemplateSpec)

type ComponentYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Subtitle   string           `yaml:"subtitle"`
	Data       *DataYAML        `yaml:"data"`
	VegaLite   *string          `yaml:"vega_lite"`
	Markdown   *string          `yaml:"markdown"`
	Image      *string          `yaml:"image"`
	Template   map[string]any   `yaml:"template"`
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

	return nil
}

// parseComponentYAML parses and validates a ComponentYAML.
// It is separated from parseComponent to allow inline creation of components from a dashboard YAML file.
func (p *Parser) parseComponentYAML(tmp *ComponentYAML) (*runtimev1.ComponentSpec, []ResourceName, error) {
	// Parse the data YAML
	var refs []ResourceName
	var resolver string
	var resolverProps *structpb.Struct
	if tmp.Data != nil {
		var err error
		resolver, resolverProps, refs, err = p.parseDataYAML(tmp.Data)
		if err != nil {
			return nil, nil, err
		}
	}

	// Discover and validate the renderer
	n := 0
	var renderer string
	var rendererProps *structpb.Struct
	if tmp.VegaLite != nil {
		n++

		var vegaLiteSpec interface{}
		if err := json.Unmarshal([]byte(*tmp.VegaLite), &vegaLiteSpec); err != nil {
			return nil, nil, errors.New(`failed to parse "vega_lite" as JSON`)
		}
		if err := vegaLiteSchema.Validate(vegaLiteSpec); err != nil {
			return nil, nil, fmt.Errorf(`failed to validate "vega_lite": %w`, err)
		}

		renderer = "vega_lite"
		rendererProps = must(structpb.NewStruct(map[string]any{"spec": *tmp.VegaLite}))
	}
	if tmp.Markdown != nil {
		n++
		renderer = "markdown"
		rendererProps = must(structpb.NewStruct(map[string]any{"content": *tmp.Markdown}))
	}
	if tmp.Image != nil {
		n++
		renderer = "image"
		rendererProps = must(structpb.NewStruct(map[string]any{"url": *tmp.Image}))
	}
	if len(tmp.Template) > 0 {
		n++

		if err := componentTemplateSchema.Validate(tmp.Template); err != nil {
			return nil, nil, fmt.Errorf(`failed to validate "template": %w`, err)
		}

		renderer = "template"
		rendererProps = must(structpb.NewStruct(tmp.Template))
	}

	// Check there is exactly one renderer
	if n == 0 {
		return nil, nil, errors.New(`missing renderer configuration (set one of vega_lite, markdown, image)`)
	}
	if n > 1 {
		return nil, nil, errors.New(`multiple renderers are not allowed (set only one of vega_lite, markdown, image)`)
	}

	// Create the component spec
	spec := &runtimev1.ComponentSpec{
		Title:              tmp.Title,
		Subtitle:           tmp.Subtitle,
		Resolver:           resolver,
		ResolverProperties: resolverProps,
		Renderer:           renderer,
		RendererProperties: rendererProps,
	}

	return spec, refs, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
