package rillv1

import (
	"bytes"
	"fmt"
	"text/template"
	"text/template/parse"

	"github.com/Masterminds/sprig/v3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

// Template parsing serves a dual purpose of:
//
// a) extracting metadata at parse time (such as {{ config ...}} and {{ ref ... }})
// b) populating values at resolve time (such as {{ .vars ... }} and {{ ref ... }})
//
// The resolve time of a template varies. For models, the resolve time is when they are created in the database.
// But for dashboard expressions, the resolve time is when the dashboard is rendered.
//
// Note that no template resolution happens at parse time. This means templating can't be used to alter the structure of YAML files.
// Instead, templating can be used to alter values in certain YAML properties at resolve time.
// This is similar to the templating behavior you would see in Github Actions, but not in Helm.
//
// The supported template functions are (if not supported at parse time or resolve time, they are no-ops resolving to empty strings):
//
//     configure `YAML`: set config from YAML blob (parse time)
//     configure `key` `value`: set config key (parse time)
//     dependency [`kind`] `name`: register a dependency (parse time)
//     ref [`kind`] `name`: register a dependency at parse-time, resolve it to a name at resolve time (parse time and resolve time)
//     lookup [`kind`] `name`: lookup another resource (resolve time)
//     .vars.name: access a variable (resolve time)
//     .user.attribute: access an attribute from auth claims (resolve time)
//     .meta: access the current resource's metadata (resolve time)
//     .spec: access the current resource's spec (resolve time)
//     .state: access the current resource's state (resolve time)
//     (All functions from Sprig except OS functions. See http://masterminds.github.io/sprig/ for details.)
//

// TemplateData contains data for resolving a template.
type TemplateData struct {
	Environment string
	User        map[string]any
	Variables   map[string]string
	State       map[string]any
	ExtraProps  map[string]any
	Self        TemplateResource
	Resolve     func(ref ResourceName) (string, error)
	Lookup      func(name ResourceName) (TemplateResource, error)
}

// TemplateResource contains data for a resource for injection into a template.
type TemplateResource struct {
	Meta  *runtimev1.ResourceMeta
	Spec  any
	State any
}

// TemplateMetadata contains metadata extracted from a template.
type TemplateMetadata struct {
	Refs                     []ResourceName
	Config                   map[string]any
	Variables                []string
	UsesTemplating           bool
	ResolvedWithPlaceholders string
}

// AnalyzeTemplate parses a template and extracts metadata.
func AnalyzeTemplate(tmpl string) (*TemplateMetadata, error) {
	// Accumulate metadata
	config := make(map[string]any)
	refs := map[ResourceName]bool{}

	// Build func map
	funcMap := newFuncMap("", nil)
	funcMap["configure"] = func(parts ...any) (string, error) {
		if len(parts) == 1 {
			// Configure from YAML
			yamlStr, ok := parts[0].(string)
			if !ok {
				return "", fmt.Errorf(`"configure" input must be a string`)
			}
			// Parse YAML into config
			err := yaml.Unmarshal([]byte(yamlStr), &config)
			if err != nil {
				return "", fmt.Errorf(`"configure" failed to parse YAML: %w`, err)
			}
			return "", nil
		} else if len(parts) == 2 {
			// Configure from key-value pair
			key, ok := parts[0].(string)
			if !ok {
				return "", fmt.Errorf(`"configure" key must be a string`)
			}
			config[key] = parts[1]
			return "", nil
		}
		return "", fmt.Errorf(`"configure" takes one or two arguments`)
	}
	funcMap["dependency"] = func(parts ...string) (string, error) {
		name, err := resourceNameFromArgs(parts...)
		if err != nil {
			return "", fmt.Errorf(`invalid "dependency" args: %w`, err)
		}
		refs[name] = true
		return "", nil
	}
	funcMap["ref"] = func(parts ...string) (string, error) {
		name, err := resourceNameFromArgs(parts...)
		if err != nil {
			return "", fmt.Errorf(`invalid "ref" args: %w`, err)
		}
		refs[name] = true
		return "<no value>", nil
	}
	funcMap["lookup"] = func(parts ...string) (map[string]any, error) {
		name, err := resourceNameFromArgs(parts...)
		if err != nil {
			return nil, fmt.Errorf(`invalid "lookup" args: %w`, err)
		}
		refs[name] = true
		return map[string]any{}, nil
	}

	// Parse template (error on missing keys)
	t, err := template.New("").Funcs(funcMap).Option("missingkey=default").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	// Build template data
	dataMap := map[string]interface{}{
		"env":   "",
		"user":  map[string]any{},
		"vars":  map[string]any{},
		"state": map[string]any{},
		"self": map[string]any{
			"meta":  map[string]any{},
			"spec":  map[string]any{},
			"state": map[string]any{},
		},
	}

	// Resolve template
	res := new(bytes.Buffer)
	if err := t.Execute(res, dataMap); err != nil {
		return nil, err
	}

	// Check if there is any templating
	noTemplating := len(t.Root.Nodes) == 0 || len(t.Root.Nodes) == 1 && t.Root.Nodes[0].Type() == parse.NodeText

	// Done
	variables := extractVariablesFromTemplate(t.Tree)
	return &TemplateMetadata{
		Refs:                     maps.Keys(refs),
		Config:                   config,
		Variables:                variables,
		UsesTemplating:           !noTemplating,
		ResolvedWithPlaceholders: res.String(),
	}, nil
}

// ResolveTemplate resolves a template to a string using the given data.
func ResolveTemplate(tmpl string, data TemplateData) (string, error) {
	// Base func map
	funcMap := newFuncMap(data.Environment, data.State)

	// Add no-ops
	funcMap["configure"] = func(parts ...string) error { return nil }
	funcMap["dependency"] = func(parts ...string) error { return nil }

	// Add func to resolve a "ref"
	funcMap["ref"] = func(parts ...string) (string, error) {
		// Parse the resource name
		name, err := resourceNameFromArgs(parts...)
		if err != nil {
			return "", fmt.Errorf(`invalid "ref" input: %w`, err)
		}

		// Resolve the ref
		ref, err := data.Resolve(name)
		if err != nil {
			return "", fmt.Errorf(`function "ref" failed: %w`, err)
		}

		// Return formatted as a map
		return ref, nil
	}

	// Add func to lookup another resource
	funcMap["lookup"] = func(parts ...string) (map[string]any, error) {
		// Support is optional
		if data.Lookup == nil {
			return nil, fmt.Errorf(`function "lookup" is not supported in this context`)
		}

		// Parse the resource name
		name, err := resourceNameFromArgs(parts...)
		if err != nil {
			return nil, fmt.Errorf(`invalid "lookup" input: %w`, err)
		}

		// Lookup the resource
		resource, err := data.Lookup(name)
		if err != nil {
			return nil, fmt.Errorf(`function "lookup" failed: %w`, err)
		}

		// Return formatted as a map
		return map[string]any{
			"meta":  resource.Meta,
			"spec":  resource.Spec,
			"state": resource.State,
		}, nil
	}

	// Parse template (error on missing keys)
	// TODO: missingkey=error may be problematic for claims.
	t, err := template.New("").Funcs(funcMap).Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Build template data
	dataMap := map[string]interface{}{
		"env":   data.Environment,
		"user":  data.User,
		"vars":  data.Variables,
		"state": data.State,
		"self": map[string]any{
			"meta":  data.Self.Meta,
			"spec":  data.Self.Spec,
			"state": data.Self.State,
		},
	}

	// Add extra props
	for k, v := range data.ExtraProps {
		dataMap[k] = v
	}

	// Resolve template
	res := new(bytes.Buffer)
	if err := t.Execute(res, dataMap); err != nil {
		return "", err
	}

	return res.String(), nil
}

// newFuncMap creates a base func map for templates.
func newFuncMap(environment string, state map[string]any) template.FuncMap {
	// Add Sprig template functions (removing functions that leak host info)
	// Derived from Helm: https://github.com/helm/helm/blob/main/pkg/engine/funcs.go
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// Add helpers for checking for common environments
	funcMap["dev"] = func() bool { return environment == "dev" }
	funcMap["prod"] = func() bool { return environment == "prod" }

	// Add helper for checking .state.incremental
	funcMap["incremental"] = func() bool { return state != nil && state["incremental"] == true }

	return funcMap
}

// resourceNameFromArgs builds a ResourceName from a list of args to a template function (currently "lookup" and "ref").
// It supports two forms: `fn "name"` or `fn "kind" "name"`
// In the first case, the Kind will be empty and upstream logic is expected to disambiguate.
func resourceNameFromArgs(parts ...string) (ResourceName, error) {
	if len(parts) == 1 {
		return ResourceName{Name: parts[0]}, nil
	}

	if len(parts) != 2 {
		return ResourceName{}, fmt.Errorf("expected one or two args, but got %d", len(parts))
	}

	kind, err := ParseResourceKind(parts[0])
	if err != nil {
		return ResourceName{}, err
	}

	return ResourceName{
		Kind: kind,
		Name: parts[1],
	}, nil
}

func EvaluateBoolExpression(expr string) (bool, error) {
	if expr == "" {
		return false, fmt.Errorf("cannot evaluate empty expression")
	}
	result, err := duckdbsql.EvaluateBool(expr)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", err)
	}
	return result, nil
}

func extractVariablesFromTemplate(tree *parse.Tree) []string {
	variablesMap := make(map[string]bool)
	walkNodes(tree.Root, func(n parse.Node) {
		if vn, ok := n.(*parse.FieldNode); ok {
			v := joinIdentifiers(vn.Ident)
			variablesMap[v] = true
		}
	})

	return maps.Keys(variablesMap)
}

func walkNodes(node parse.Node, fn func(n parse.Node)) {
	fn(node)
	switch n := node.(type) {
	case *parse.ListNode:
		for _, ln := range n.Nodes {
			walkNodes(ln, fn)
		}
	case *parse.ActionNode:
		walkNodes(n.Pipe, fn)
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			walkNodes(cmd, fn)
		}
	case *parse.CommandNode:
		for _, arg := range n.Args {
			walkNodes(arg, fn)
		}
	default:
		return
	}
}

func joinIdentifiers(ident []string) string {
	var result string
	for _, id := range ident {
		if result != "" {
			result += "."
		}
		result += id
	}
	return result
}
