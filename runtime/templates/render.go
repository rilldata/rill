package templates

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

// RenderInput contains all parameters for rendering a template.
type RenderInput struct {
	Template      *Template
	Output        string         // "connector", "model", or "" for all files
	Properties    map[string]any // raw form values
	ConnectorName string         // for model outputs: the connector to reference
	ExistingEnv   map[string]bool
}

// RenderOutput contains the result of rendering.
type RenderOutput struct {
	Files   []RenderedFile
	EnvVars map[string]string
}

// RenderedFile is a single rendered output file.
type RenderedFile struct {
	Path string
	Blob string
}

// Render executes a template with the given input, producing rendered files and env vars.
// The rendering pipeline:
// 1. Pre-processes properties (secret extraction, empty filtering, derived fields)
// 2. Builds a template data map
// 3. Renders each matching file's path and code templates
func Render(input *RenderInput) (*RenderOutput, error) {
	if input.Template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	envVars := make(map[string]string)
	existingEnv := cloneEnvMap(input.ExistingEnv)

	// Build the template data context
	data := buildTemplateData(input, existingEnv, envVars)

	// Render each matching file
	var files []RenderedFile
	for _, f := range input.Template.Files {
		if input.Output != "" && f.Name != input.Output {
			continue
		}

		path, err := renderString(f.Name+"_path", f.PathTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("rendering path template for %q: %w", f.Name, err)
		}

		blob, err := renderString(f.Name+"_code", f.CodeTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("rendering code template for %q: %w", f.Name, err)
		}

		files = append(files, RenderedFile{
			Path: strings.TrimSpace(path),
			Blob: blob,
		})
	}

	return &RenderOutput{Files: files, EnvVars: envVars}, nil
}

// buildTemplateData creates the data map passed to Go templates.
// It pre-processes properties: extracts secrets, filters empties, adds derived fields.
func buildTemplateData(input *RenderInput, existingEnv map[string]bool, envVars map[string]string) map[string]any {
	data := make(map[string]any)

	// Basic fields
	data["driver"] = input.Template.Driver
	data["connector_name"] = input.ConnectorName
	data["docs_url"] = input.Template.DocsURL

	// Derive model_name from the "name" property if present
	if name, ok := input.Properties["name"]; ok && !isEmpty(name) {
		data["model_name"] = fmt.Sprintf("%v", name)
	}

	// Pre-populate all schema properties with empty strings so the base YAML
	// skeleton renders cleanly even before the user fills in any values.
	if input.Template.JSONSchema != nil {
		for k := range schemaProperties(input.Template.JSONSchema) {
			data[k] = ""
		}
	}

	// Copy all raw properties into data (overwrites empty defaults above)
	for k, v := range input.Properties {
		if !isEmpty(v) {
			data[k] = fmt.Sprintf("%v", v)
		}
	}

	// Schema-based path: extract property metadata from JSON Schema
	if input.Template.JSONSchema != nil {
		allProps := processPropertiesFromSchema(
			input.Template.JSONSchema,
			input.Template.PropertyOrder,
			input.Template.Driver,
			input.Properties,
			existingEnv,
			envVars,
		)

		// Split by x-step so connector and model files get the right properties
		configProps, sourceProps := splitPropsByStep(allProps, input.Template.JSONSchema)
		data["props"] = configProps
		data["config_props"] = configProps
		data["source_props"] = sourceProps

		return data
	}

	// Template without schema: pass properties as-is
	return data
}

// processPropertiesFromSchema pre-processes properties using JSON Schema metadata.
// Fields with "x-secret": true are extracted to env vars; "x-ui-only": true fields are skipped.
// Quoting is determined by the schema "type" field (number/boolean unquoted; everything else quoted).
func processPropertiesFromSchema(
	schema map[string]any,
	propertyOrder []string,
	driverName string,
	rawProps map[string]any,
	existingEnv map[string]bool,
	envVars map[string]string,
) []ProcessedProp {
	propsMap := schemaProperties(schema)
	if len(propsMap) == 0 {
		return nil
	}

	// Use schema-defined property order if available (preserves JSON key ordering);
	// fall back to alphabetical for deterministic output.
	keys := propertyOrder
	if len(keys) == 0 {
		keys = make([]string, 0, len(propsMap))
		for k := range propsMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}

	var result []ProcessedProp
	for _, key := range keys {
		prop := propsMap[key]

		// Skip UI-only fields (e.g. auth_method radio buttons)
		if schemaFieldBool(prop, "x-ui-only") {
			continue
		}

		val, ok := rawProps[key]
		if !ok || isEmpty(val) {
			continue
		}

		// Skip managed: false for ClickHouse (it's the default)
		if key == "managed" && !toBool(val) {
			continue
		}

		// Handle map-typed properties (e.g. headers).
		// The frontend key-value editor sends [{key, value}, ...]; convert to a flat map.
		mapVal, isMap := val.(map[string]any)
		if !isMap {
			if arrVal, isArr := val.([]any); isArr {
				mapVal = kvArrayToMap(arrVal)
				isMap = mapVal != nil
			}
		}
		if isMap {
			headerIdent := driverName
			headerProps := processHeaders(mapVal, headerIdent, existingEnv, envVars)
			result = append(result, headerProps...)
			continue
		}

		strVal := fmt.Sprintf("%v", val)

		if schemaFieldBool(prop, "x-secret") {
			envName := ResolveEnvVarNameForKey(driverName, key, schemaFieldString(prop, "x-env-var"), existingEnv)
			existingEnv[envName] = true
			envVars[envName] = strVal
			result = append(result, ProcessedProp{
				Key:    key,
				Value:  fmt.Sprintf("{{ .env.%s }}", envName),
				Quoted: true,
			})
		} else {
			propType := schemaFieldString(prop, "type")
			quoted := propType != "number" && propType != "boolean"
			result = append(result, ProcessedProp{
				Key:    key,
				Value:  strVal,
				Quoted: quoted,
			})
		}
	}
	return result
}

// splitPropsByStep separates processed props into connector-step and source-step slices
// based on the x-step field in the JSON Schema. Props without x-step go into both.
func splitPropsByStep(props []ProcessedProp, schema map[string]any) (configProps, sourceProps []ProcessedProp) {
	propsMap := schemaProperties(schema)
	for _, p := range props {
		step := ""
		if propSchema, ok := propsMap[p.Key]; ok {
			step = schemaFieldString(propSchema, "x-step")
		}
		switch step {
		case "connector":
			configProps = append(configProps, p)
		case "source":
			sourceProps = append(sourceProps, p)
		case "explorer":
			// Explorer-step props are accessed directly as template variables (e.g. .sql, .name);
			// they are excluded from renderProps output to avoid duplication.
		default:
			configProps = append(configProps, p)
			sourceProps = append(sourceProps, p)
		}
	}
	return
}

// kvArrayToMap converts [{key: "k", value: "v"}, ...] (from the frontend key-value editor)
// to map[string]any{"k": "v", ...}. Returns nil if the array is empty or not in the expected format.
func kvArrayToMap(arr []any) map[string]any {
	result := make(map[string]any, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		k, _ := obj["key"].(string)
		v, _ := obj["value"].(string)
		if k != "" {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// schemaProperties extracts the "properties" map from a JSON Schema object.
func schemaProperties(schema map[string]any) map[string]map[string]any {
	raw, ok := schema["properties"]
	if !ok {
		return nil
	}
	outer, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]map[string]any, len(outer))
	for k, v := range outer {
		if prop, ok := v.(map[string]any); ok {
			result[k] = prop
		}
	}
	return result
}

// schemaFieldBool returns the bool value of a field in a schema property map.
func schemaFieldBool(prop map[string]any, key string) bool {
	v, _ := prop[key].(bool)
	return v
}

// schemaFieldString returns the string value of a field in a schema property map.
func schemaFieldString(prop map[string]any, key string) string {
	v, _ := prop[key].(string)
	return v
}

// processHeaders processes a map of header key-value pairs, extracting sensitive values.
func processHeaders(headers map[string]any, connectorName string, existingEnv map[string]bool, envVars map[string]string) []ProcessedProp {
	if len(headers) == 0 {
		return nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build a single "headers" prop whose value is a YAML mapping
	var headerLines []string
	for _, key := range keys {
		strVal := fmt.Sprintf("%v", headers[key])
		if IsSensitiveHeaderKey(key) {
			envSegment := HeaderKeyToEnvSegment(key)
			envName := ResolveHeaderEnvVarName(connectorName, envSegment, existingEnv)
			existingEnv[envName] = true
			if scheme, secret, ok := SplitAuthSchemePrefix(strVal); ok {
				envVars[envName] = secret
				headerLines = append(headerLines, fmt.Sprintf("  %s: \"%s{{ .env.%s }}\"", key, scheme, envName))
			} else {
				envVars[envName] = strVal
				headerLines = append(headerLines, fmt.Sprintf("  %s: \"{{ .env.%s }}\"", key, envName))
			}
		} else {
			headerLines = append(headerLines, fmt.Sprintf("  %s: %q", key, strVal))
		}
	}

	// Return as a single rendered block (the template will output it as-is)
	return []ProcessedProp{{
		Key:    "headers",
		Value:  "\n" + strings.Join(headerLines, "\n"),
		Quoted: false, // The value is a nested YAML mapping, not a scalar
	}}
}

// renderString renders a Go template string using [[ ]] delimiters.
func renderString(name, tmplText string, data map[string]any) (string, error) {
	t, err := template.New(name).
		Delims("[[", "]]").
		Funcs(funcMap()).
		Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}

// isEmpty checks if a value is empty (nil, empty string, or empty map).
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case bool:
		return false // bools are never "empty"
	case map[string]any:
		return len(val) == 0
	default:
		return fmt.Sprintf("%v", v) == ""
	}
}

// toBool converts a value to bool.
func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true"
	default:
		return false
	}
}

// cloneEnvMap creates a shallow copy of an env key map.
func cloneEnvMap(m map[string]bool) map[string]bool {
	if m == nil {
		return make(map[string]bool)
	}
	clone := make(map[string]bool, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}
