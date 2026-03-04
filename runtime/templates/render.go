package templates

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/rilldata/rill/runtime/drivers"
)

// RenderInput contains all parameters for rendering a template.
type RenderInput struct {
	Template      *Template
	Output        string         // "connector", "model", or "" for all files
	DriverSpec    *drivers.Spec  // driver metadata; nil for driverless templates
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
	data, err := buildTemplateData(input, existingEnv, envVars)
	if err != nil {
		return nil, err
	}

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
func buildTemplateData(input *RenderInput, existingEnv map[string]bool, envVars map[string]string) (map[string]any, error) {
	data := make(map[string]any)

	// Basic fields
	data["driver"] = input.Template.Driver
	data["connector_name"] = input.ConnectorName

	// Derive model_name from the "name" property if present
	if name, ok := input.Properties["name"]; ok && !isEmpty(name) {
		data["model_name"] = fmt.Sprintf("%v", name)
	}

	// Copy all raw properties into data (pre-processed values will overwrite)
	for k, v := range input.Properties {
		if !isEmpty(v) {
			data[k] = fmt.Sprintf("%v", v)
		}
	}

	// Schema-based path: extract property metadata from JSON Schema
	if input.Template.JSONSchema != nil {
		allProps := processPropertiesFromSchema(
			input.Template.JSONSchema,
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

		// Compute derived fields (SQL, create_secrets_from_connectors) based on template metadata
		if input.Template.OLAP == "duckdb" && input.Template.Driver != "" {
			if err := applyDuckDBDerivedFieldsForSchema(input, data); err != nil {
				return nil, err
			}
		}

		// ClickHouse rewrite: compute SQL from properties using ClickHouse table functions
		if input.Template.OLAP == "clickhouse" && input.Template.Driver != "" {
			if err := applyClickHouseDerivedFields(input, data, configProps); err != nil {
				return nil, err
			}
		}

		return data, nil
	}

	if input.DriverSpec == nil {
		// Driverless template without schema: pass properties as-is
		return data, nil
	}

	// Process config properties (for connector outputs)
	configProps := processProperties(input.Template.Driver, input.DriverSpec.ConfigProperties, input.Properties, existingEnv, envVars, input.ConnectorName)
	data["props"] = configProps
	data["config_props"] = configProps

	// Process source properties (for model outputs)
	sourceProps := processProperties(input.Template.Driver, input.DriverSpec.SourceProperties, input.Properties, existingEnv, envVars, input.ConnectorName)
	data["source_props"] = sourceProps

	// DuckDB rewrite: compute SQL from path for object store, file store, and sqlite drivers
	if input.Template.OLAP == "duckdb" && input.Template.Driver != "" {
		if err := applyDuckDBDerivedFields(input, data); err != nil {
			return nil, err
		}
	}

	// ClickHouse rewrite: compute SQL from properties using ClickHouse table functions
	if input.Template.OLAP == "clickhouse" && input.Template.Driver != "" {
		if err := applyClickHouseDerivedFields(input, data, configProps); err != nil {
			return nil, err
		}
	}

	// Special flags
	data["no_dev"] = input.Template.Driver == "redshift"
	data["materialize"] = input.Template.Driver != "duckdb" && input.Template.Driver != "motherduck"

	return data, nil
}

// processProperties pre-processes a list of PropertySpecs against raw form values.
// For each property: filters empties, extracts secrets to env vars, and determines formatting.
func processProperties(
	driverName string,
	specs []*drivers.PropertySpec,
	rawProps map[string]any,
	existingEnv map[string]bool,
	envVars map[string]string,
	connectorName string,
) []ProcessedProp {
	var result []ProcessedProp
	for _, spec := range specs {
		val, ok := rawProps[spec.Key]
		if !ok || isEmpty(val) {
			continue
		}

		// Handle map-typed properties (e.g. headers)
		if mapVal, isMap := val.(map[string]any); isMap {
			// Use connector name for header env var naming; fall back to driver name
			headerIdent := connectorName
			if headerIdent == "" {
				headerIdent = driverName
			}
			headerProps := processHeaders(mapVal, headerIdent, existingEnv, envVars)
			result = append(result, headerProps...)
			continue
		}

		// Skip managed: false for ClickHouse (it's the default)
		if spec.Key == "managed" && !toBool(val) {
			continue
		}

		strVal := fmt.Sprintf("%v", val)
		if spec.Secret {
			envName := ResolveEnvVarName(driverName, spec, existingEnv)
			existingEnv[envName] = true
			envVars[envName] = strVal
			result = append(result, ProcessedProp{
				Key:    spec.Key,
				Value:  fmt.Sprintf("{{ .env.%s }}", envName),
				Quoted: true,
			})
		} else {
			quoted := spec.Type != drivers.NumberPropertyType && spec.Type != drivers.BooleanPropertyType
			result = append(result, ProcessedProp{
				Key:    spec.Key,
				Value:  strVal,
				Quoted: quoted,
			})
		}
	}
	return result
}

// processPropertiesFromSchema pre-processes properties using JSON Schema metadata instead of drivers.PropertySpec.
// Fields with "x-secret": true are extracted to env vars; "x-ui-only": true fields are skipped.
// Quoting is determined by the schema "type" field (number/boolean unquoted; everything else quoted).
func processPropertiesFromSchema(
	schema map[string]any,
	driverName string,
	rawProps map[string]any,
	existingEnv map[string]bool,
	envVars map[string]string,
) []ProcessedProp {
	propsMap := schemaProperties(schema)
	if len(propsMap) == 0 {
		return nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(propsMap))
	for k := range propsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

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

		// Handle map-typed properties (e.g. headers)
		if mapVal, isMap := val.(map[string]any); isMap {
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

// applyDuckDBDerivedFieldsForSchema computes DuckDB-specific derived fields for schema-based templates.
// Uses the template's driver name and JSON Schema x-category instead of DriverSpec.
func applyDuckDBDerivedFieldsForSchema(input *RenderInput, data map[string]any) error {
	category := schemaFieldString(input.Template.JSONSchema, "x-category")

	// Special case for sqlite: uses db+table instead of path
	if input.Template.Driver == "sqlite" {
		db := strVal(input.Properties["db"])
		table := strVal(input.Properties["table"])
		if db == "" || table == "" {
			return nil
		}
		data["sql"] = fmt.Sprintf("SELECT * FROM sqlite_scan('%s', '%s');", db, table)
		return nil
	}

	path := strVal(input.Properties["path"])

	// Path may be empty when rendering only the connector output; skip derivation.
	if path == "" {
		return nil
	}

	if input.ConnectorName != "" {
		data["create_secrets_from_connectors"] = input.ConnectorName
	}

	switch {
	case category == "objectStore": // s3, gcs, azure
		data["sql"] = BuildDuckDBQuery(path, false)
	case category == "fileStore" || input.Template.Driver == "https":
		data["sql"] = BuildDuckDBQuery(path, true)
	}
	return nil
}

// applyDuckDBDerivedFields computes DuckDB-specific derived fields (SQL, create_secrets_from_connectors).
func applyDuckDBDerivedFields(input *RenderInput, data map[string]any) error {
	spec := input.DriverSpec
	path := strVal(input.Properties["path"])

	switch {
	case spec.ImplementsObjectStore: // s3, gcs, azure
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for %s DuckDB model", input.Template.Driver)
		}
		if input.ConnectorName != "" {
			data["create_secrets_from_connectors"] = input.ConnectorName
		}
		data["sql"] = BuildDuckDBQuery(path, false)

	case input.Template.Driver == "https":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for HTTPS DuckDB model")
		}
		if input.ConnectorName != "" {
			data["create_secrets_from_connectors"] = input.ConnectorName
		}
		data["sql"] = BuildDuckDBQuery(path, true)

	case spec.ImplementsFileStore: // local_file
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for local file DuckDB model")
		}
		data["sql"] = BuildDuckDBQuery(path, false)

	case input.Template.Driver == "sqlite":
		db := strVal(input.Properties["db"])
		table := strVal(input.Properties["table"])
		if db == "" || table == "" {
			return fmt.Errorf("missing required properties \"db\" and \"table\" for SQLite DuckDB model")
		}
		data["sql"] = fmt.Sprintf("SELECT * FROM sqlite_scan('%s', '%s');", db, table)
	}
	return nil
}

// applyClickHouseDerivedFields computes ClickHouse-specific SQL using native table functions.
// ClickHouse models embed credentials directly in the SQL as env var references.
func applyClickHouseDerivedFields(input *RenderInput, data map[string]any, configProps []ProcessedProp) error {
	path := strVal(input.Properties["path"])

	// Build a lookup of processed config prop values by key (env var refs for secrets)
	propVal := make(map[string]string, len(configProps))
	for _, p := range configProps {
		propVal[p.Key] = p.Value
	}

	switch input.Template.Driver {
	case "s3":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for S3 ClickHouse model")
		}
		data["sql"] = BuildClickHouseObjectStoreQuery("s3", path,
			propVal["aws_access_key_id"], propVal["aws_secret_access_key"])

	case "gcs":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for GCS ClickHouse model")
		}
		data["sql"] = BuildClickHouseObjectStoreQuery("gcs", path,
			propVal["key_id"], propVal["secret"])

	case "azure":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for Azure ClickHouse model")
		}
		container, blobPath := parseAzurePath(path)
		endpoint := fmt.Sprintf("https://%s.blob.core.windows.net",
			strVal(input.Properties["azure_storage_account"]))
		data["sql"] = BuildClickHouseAzureQuery(endpoint, container, blobPath,
			propVal["azure_storage_account"], propVal["azure_storage_key"])

	case "mysql":
		host := strVal(input.Properties["host"])
		if host == "" {
			return fmt.Errorf("missing required property \"host\" for MySQL ClickHouse model")
		}
		port := strVal(input.Properties["port"])
		if port == "" {
			port = "3306"
		}
		data["sql"] = BuildClickHouseDatabaseQuery("mysql", host+":"+port,
			strVal(input.Properties["database"]),
			strVal(input.Properties["table"]),
			propVal["user"], propVal["password"])

	case "postgres":
		host := strVal(input.Properties["host"])
		if host == "" {
			return fmt.Errorf("missing required property \"host\" for Postgres ClickHouse model")
		}
		port := strVal(input.Properties["port"])
		if port == "" {
			port = "5432"
		}
		data["sql"] = BuildClickHouseDatabaseQuery("postgresql", host+":"+port,
			strVal(input.Properties["dbname"]),
			strVal(input.Properties["table"]),
			propVal["user"], propVal["password"])

	case "https":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for HTTPS ClickHouse model")
		}
		data["sql"] = BuildClickHouseURLQuery(path)

	case "local_file":
		if path == "" {
			return fmt.Errorf("missing required property \"path\" for local file ClickHouse model")
		}
		data["sql"] = BuildClickHouseFileQuery(path)

	case "sqlite":
		db := strVal(input.Properties["db"])
		table := strVal(input.Properties["table"])
		if db == "" || table == "" {
			return fmt.Errorf("missing required properties \"db\" and \"table\" for SQLite ClickHouse model")
		}
		data["sql"] = BuildClickHouseSQLiteQuery(db, table)
	}
	return nil
}

// parseAzurePath parses "azure://container/blob/path" into container and blob path.
func parseAzurePath(path string) (container, blobPath string) {
	path = strings.TrimPrefix(path, "azure://")
	idx := strings.IndexByte(path, '/')
	if idx < 0 {
		return path, ""
	}
	return path[:idx], path[idx+1:]
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

// strVal extracts a string value from an interface.
func strVal(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
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
