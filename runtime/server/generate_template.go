package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// GenerateTemplate generates a connector or model YAML file from structured form data.
func (s *Server) GenerateTemplate(ctx context.Context, req *runtimev1.GenerateTemplateRequest) (*runtimev1.GenerateTemplateResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.resource_type", req.ResourceType),
		attribute.String("args.driver", req.Driver),
		attribute.String("args.connector_name", req.ConnectorName),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	// Validate resource type
	if req.ResourceType != "connector" && req.ResourceType != "model" {
		return nil, status.Errorf(codes.InvalidArgument, "resource_type must be \"connector\" or \"model\"")
	}

	// Validate driver exists
	drv, ok := drivers.Connectors[req.Driver]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown driver %q", req.Driver)
	}
	spec := drv.Spec()

	// Convert properties
	var props map[string]any
	if req.Properties != nil {
		props = req.Properties.AsMap()
	} else {
		props = make(map[string]any)
	}

	// Validate properties against the original driver spec (before any rewrite)
	if err := validateProperties(spec, req.ResourceType, props); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	// DuckDB rewrite for object store, file store, and sqlite drivers
	actualDriver := req.Driver
	actualResourceType := req.ResourceType
	if req.ResourceType == "model" {
		actualDriver, props = maybeRewriteToDuckDB(spec, req.Driver, props, req.ConnectorName)
	}

	// Read existing .env for env var conflict resolution
	existingEnv := make(map[string]bool)
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err == nil {
		existingEnv = readEnvKeys(ctx, repo)
		release()
	}

	// Render YAML + extract secrets
	var blob string
	envVars := make(map[string]string)
	switch actualResourceType {
	case "connector":
		blob = renderConnectorYAML(spec, actualDriver, props, existingEnv, envVars)
	case "model":
		if actualDriver == req.Driver {
			// Non-rewritten model (warehouse driver)
			blob = renderModelYAML(spec, actualDriver, props, req.ConnectorName, existingEnv, envVars)
		} else {
			// Rewritten to DuckDB
			blob = renderDuckDBModelYAML(props, req.ConnectorName)
		}
	}

	return &runtimev1.GenerateTemplateResponse{
		Blob:         blob,
		EnvVars:      envVars,
		ResourceType: actualResourceType,
		Driver:       actualDriver,
	}, nil
}

// validateProperties rejects unknown property keys.
func validateProperties(spec drivers.Spec, resourceType string, properties map[string]any) error {
	allowed := make(map[string]bool)
	var props []*drivers.PropertySpec
	if resourceType == "connector" {
		props = spec.ConfigProperties
	} else {
		props = spec.SourceProperties
		// Universal model properties: sql and name are always allowed for models,
		// even if the driver's SourceProperties doesn't list them explicitly.
		allowed["sql"] = true
		allowed["name"] = true
	}
	for _, p := range props {
		allowed[p.Key] = true
	}
	for key := range properties {
		if !allowed[key] {
			return fmt.Errorf("unknown property %q for driver", key)
		}
	}
	return nil
}

// maybeRewriteToDuckDB transforms object store, file store, and sqlite drivers into DuckDB models.
func maybeRewriteToDuckDB(spec drivers.Spec, driverName string, props map[string]any, connectorName string) (string, map[string]any) {
	if !spec.ImplementsObjectStore && !spec.ImplementsFileStore && driverName != "sqlite" {
		return driverName, props
	}

	rewritten := make(map[string]any, len(props))
	for k, v := range props {
		rewritten[k] = v
	}

	switch {
	case spec.ImplementsObjectStore: // s3, gcs, azure
		if connectorName != "" {
			rewritten["create_secrets_from_connectors"] = connectorName
		}
		rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), false)
		delete(rewritten, "path")
		delete(rewritten, "name")

	case driverName == "https":
		if connectorName != "" {
			rewritten["create_secrets_from_connectors"] = connectorName
		}
		rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), true)
		delete(rewritten, "path")
		delete(rewritten, "name")

	case spec.ImplementsFileStore: // local_file
		rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), false)
		delete(rewritten, "path")
		delete(rewritten, "name")
		delete(rewritten, "format")

	case driverName == "sqlite":
		rewritten["sql"] = fmt.Sprintf("SELECT * FROM sqlite_scan('%s', '%s');", strVal(props["db"]), strVal(props["table"]))
		delete(rewritten, "db")
		delete(rewritten, "table")
		delete(rewritten, "name")
	}

	return "duckdb", rewritten
}

// buildDuckDBQuery maps a file path + extension to a DuckDB read function.
func buildDuckDBQuery(path string, defaultToJSON bool) string {
	ext := strings.ToLower(filepath.Ext(path))
	// Handle compound extensions like .v1.parquet.gz by checking if ext is contained
	fullLower := strings.ToLower(path)
	switch {
	case containsExt(fullLower, ext, ".csv", ".tsv", ".txt"):
		return fmt.Sprintf("select * from read_csv('%s', auto_detect=true, ignore_errors=1, header=true)", path)
	case containsExt(fullLower, ext, ".parquet"):
		return fmt.Sprintf("select * from read_parquet('%s')", path)
	case containsExt(fullLower, ext, ".json", ".ndjson"):
		return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
	default:
		if defaultToJSON {
			return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
		}
		return fmt.Sprintf("select * from '%s'", path)
	}
}

// containsExt checks if the file extension or full path contains any of the target extensions.
// This handles compound extensions like .v1.parquet.gz.
func containsExt(fullLower, ext string, targets ...string) bool {
	for _, t := range targets {
		if ext == t {
			return true
		}
		// Check for compound extensions
		if strings.Contains(fullLower, t) {
			return true
		}
	}
	return false
}

// renderConnectorYAML builds a connector YAML file using yaml.Node for precise formatting.
func renderConnectorYAML(spec drivers.Spec, driverName string, props map[string]any, existingEnv map[string]bool, envVars map[string]string) string {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	mapping.HeadComment = fmt.Sprintf("Connector YAML\nReference documentation: %s", spec.DocsURL)

	addScalarPair(mapping, "type", "connector")
	addScalarPair(mapping, "driver", driverName)

	for _, propSpec := range spec.ConfigProperties {
		val, ok := props[propSpec.Key]
		if !ok || isEmpty(val) {
			continue
		}
		// Skip managed: false for ClickHouse (it's the default)
		if propSpec.Key == "managed" && !toBool(val) {
			continue
		}
		if propSpec.Secret {
			envName := resolveEnvVarName(driverName, propSpec, existingEnv)
			existingEnv[envName] = true
			envVars[envName] = fmt.Sprintf("%v", val)
			addQuotedPair(mapping, propSpec.Key, fmt.Sprintf("{{ .env.%s }}", envName))
		} else {
			addTypedPair(mapping, propSpec, val)
		}
	}

	doc.Content = append(doc.Content, mapping)
	return encodeYAML(doc)
}

// renderModelYAML builds a model YAML file for warehouse drivers (non-rewritten).
func renderModelYAML(spec drivers.Spec, driverName string, props map[string]any, connectorName string, existingEnv map[string]bool, envVars map[string]string) string {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	mapping.HeadComment = fmt.Sprintf("Model YAML\nReference documentation: %s", spec.DocsURL)

	addScalarPair(mapping, "type", "model")
	if connectorName != "" {
		addQuotedPair(mapping, "connector", connectorName)
	}

	// Non-DuckDB warehouse models need materialize: true
	if driverName != "duckdb" && driverName != "motherduck" {
		addScalarPair(mapping, "materialize", "true")
	}

	// Add source properties from the driver spec
	sqlHandled := false
	for _, propSpec := range spec.SourceProperties {
		val, ok := props[propSpec.Key]
		if !ok || isEmpty(val) {
			continue
		}
		if propSpec.Key == "name" {
			continue // name is used for the file path, not in YAML
		}
		if propSpec.Key == "sql" {
			addSQLBlock(mapping, fmt.Sprintf("%v", val))
			sqlHandled = true
			continue
		}
		if propSpec.Secret {
			envName := resolveEnvVarName(driverName, propSpec, existingEnv)
			existingEnv[envName] = true
			envVars[envName] = fmt.Sprintf("%v", val)
			addQuotedPair(mapping, propSpec.Key, fmt.Sprintf("{{ .env.%s }}", envName))
		} else {
			addTypedPair(mapping, propSpec, val)
		}
	}

	// Handle sql as a universal model property (warehouse drivers don't list it in SourceProperties)
	if !sqlHandled {
		if sql, ok := props["sql"]; ok && !isEmpty(sql) {
			addSQLBlock(mapping, fmt.Sprintf("%v", sql))
		}
	}

	// Dev section with limit (except Redshift)
	if driverName != "redshift" {
		addDevSection(mapping)
	}

	doc.Content = append(doc.Content, mapping)
	return encodeYAML(doc)
}

// renderDuckDBModelYAML builds a DuckDB model YAML file for rewritten object/file store drivers.
func renderDuckDBModelYAML(props map[string]any, connectorName string) string {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	mapping.HeadComment = "Model YAML\nReference documentation: https://docs.rilldata.com/developers/build/connectors/olap/duckdb"

	addScalarPair(mapping, "type", "model")
	addQuotedPair(mapping, "connector", "duckdb")

	// Add create_secrets_from_connectors if present
	if secretsConn, ok := props["create_secrets_from_connectors"]; ok && !isEmpty(secretsConn) {
		addQuotedPair(mapping, "create_secrets_from_connectors", fmt.Sprintf("%v", secretsConn))
	}

	// Add SQL block
	if sql, ok := props["sql"]; ok && !isEmpty(sql) {
		addSQLBlock(mapping, fmt.Sprintf("%v", sql))
	}

	doc.Content = append(doc.Content, mapping)
	return encodeYAML(doc)
}

// resolveEnvVarName determines the env var name for a secret property, resolving conflicts.
func resolveEnvVarName(driverName string, propSpec *drivers.PropertySpec, existingEnv map[string]bool) string {
	var base string
	if propSpec.EnvVarName != "" {
		base = propSpec.EnvVarName
	} else {
		// Fallback: DRIVER_KEY format (SCREAMING_SNAKE_CASE)
		base = strings.ToUpper(driverName) + "_" + strings.ToUpper(propSpec.Key)
	}

	// Check for conflicts
	candidate := base
	for i := 1; existingEnv[candidate]; i++ {
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return candidate
}

// readEnvKeys parses an .env file into a set of key names.
func readEnvKeys(ctx context.Context, repo drivers.RepoStore) map[string]bool {
	keys := make(map[string]bool)
	content, err := repo.Get(ctx, ".env")
	if err != nil {
		return keys
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			keys[line[:idx]] = true
		}
	}
	return keys
}

// addScalarPair adds a key-value pair with plain scalar style.
func addScalarPair(m *yaml.Node, key, value string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}

// addQuotedPair adds a key-value pair with double-quoted value style.
func addQuotedPair(m *yaml.Node, key, value string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Style: yaml.DoubleQuotedStyle},
	)
}

// addTypedPair adds a key-value pair with appropriate formatting based on property type.
func addTypedPair(m *yaml.Node, propSpec *drivers.PropertySpec, val any) {
	strVal := fmt.Sprintf("%v", val)
	switch propSpec.Type {
	case drivers.NumberPropertyType, drivers.BooleanPropertyType:
		addScalarPair(m, propSpec.Key, strVal)
	default:
		addQuotedPair(m, propSpec.Key, strVal)
	}
}

// addSQLBlock adds a SQL key with literal block style (|).
func addSQLBlock(m *yaml.Node, sql string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "sql"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: sql, Style: yaml.LiteralStyle},
	)
}

// addDevSection adds a dev section with a SQL limit clause.
func addDevSection(m *yaml.Node) {
	devMapping := &yaml.Node{Kind: yaml.MappingNode}
	addSQLBlock(devMapping, "select * from {{ ref \"self\" }} limit 10000")
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "dev"},
		devMapping,
	)
}

// encodeYAML renders a yaml.Node tree to a string.
func encodeYAML(doc *yaml.Node) string {
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return ""
	}
	if err := enc.Close(); err != nil {
		return ""
	}
	return buf.String()
}

// isEmpty checks if a value is empty (nil, empty string, or false for booleans).
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case bool:
		return false // bools are never "empty"; handled explicitly where needed
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
