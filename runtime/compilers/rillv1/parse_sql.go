package rillv1

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// sqlConfigYAML contains YAML fields that provide context for parsing of a .sql file.
// It must be present at the same path as the .sql file (i.e. "/path/to/file.yaml" for "/path/to/file.sql").
type sqlConfigYAML struct {
	genericYAML  `yaml:",inline"`
	Connector    string `yaml:"connector"`
	ParserConfig struct {
		DisableTemplating            bool `yaml:"disable_templating"`
		DisableDuckDBInference       bool `yaml:"disable_duckdb_inference"`
		DisableDuckDBSourceRewriting bool `yaml:"disable_duckdb_source_rewriting"`
	} `yaml:"parser"`
}

// sqlConfig contains generic information about a SQL file derived from sqlConfigYAML and from templating inside the SQL.
type sqlConfig struct {
	Kind                         ResourceKind
	Name                         string
	Refs                         []ResourceName
	Connector                    string
	Annotations                  map[string]any
	UsesTemplating               bool
	DisableTemplating            bool
	DisableDuckDBInference       bool
	DisableDuckDBSourceRewriting bool
}

// parseSQL parses a SQL file and adds the resulting resource(s) to p.Resources.
func (p *Parser) parseSQL(ctx context.Context, path, data string) error {
	// Config relating to the SQL file can be provided by several means. We incrementally build the SQL config.
	cfg := sqlConfig{Annotations: make(map[string]any)}
	cfg.Name = fileutil.Stem(path)

	// We treat the "sources", "models", and "dashboards" directories and "init.sql" file as providing special context.
	// Files outside must specify a "kind" in a SQL annotation or template config.
	if strings.HasPrefix(path, "/sources") {
		cfg.Kind = ResourceKindSource
	} else if strings.HasPrefix(path, "/models") {
		cfg.Kind = ResourceKindModel
	} else if strings.HasPrefix(path, "/dashboards") {
		cfg.Kind = ResourceKindMetricsView
	} else if path == "/init.sql" {
		cfg.Kind = ResourceKindMigration
	}

	// A YAML file with the same path as the SQL file can provide additional config.
	// This config could also have been parsed in the parseYAML functions and propagated to here, but it's much simpler just to read the file twice.
	yamlPath := strings.TrimSuffix(path, ".sql") + ".yaml"
	yamlData, err := p.Repo.Get(ctx, p.InstanceID, yamlPath)
	if err != nil {
		// Try .yml file instead of .yaml
		yamlPath := strings.TrimSuffix(path, ".sql") + ".yml"
		yamlData, err = p.Repo.Get(ctx, p.InstanceID, yamlPath)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading companion YAML file at %q: %w", yamlPath, err)
	}

	// Parse YAML (if found) into cfg
	if yamlData != "" {
		var tmp sqlConfigYAML
		if err := yaml.Unmarshal([]byte(yamlData), &tmp); err != nil {
			return fmt.Errorf("error parsing companion .yaml file at %q: %w", yamlPath, newYAMLError(err))
		}

		cfg.Connector = tmp.Connector

		if tmp.Name != "" {
			cfg.Name = tmp.Name
		}

		if tmp.Kind != nil {
			kind, err := ParseResourceKind(*tmp.Kind)
			if err != nil {
				return err
			}
			cfg.Kind = kind
		}

		cfg.Refs, err = parseYAMLRefs(tmp.Refs)
		if err != nil {
			return err
		}

		cfg.DisableTemplating = tmp.ParserConfig.DisableTemplating
		cfg.DisableDuckDBInference = tmp.ParserConfig.DisableDuckDBInference
		cfg.DisableDuckDBSourceRewriting = tmp.ParserConfig.DisableDuckDBSourceRewriting
	}

	// Lastly, we try extracting info through templating
	if !cfg.DisableTemplating {
		meta, err := AnalyzeTemplate(data)
		if err != nil {
			return err
		}

		cfg.UsesTemplating = meta.UsesTemplating

		// You can override connector and kind in the template config (but not other fields in sqlConfig).
		for k, v := range meta.Config {
			if strings.EqualFold(k, "connector") {
				c, ok := v.(string)
				if !ok {
					return fmt.Errorf("invalid connector <%v>", v)
				}
				cfg.Connector = c
				continue
			}

			if strings.EqualFold(k, "kind") {
				k, ok := v.(string)
				if !ok {
					return fmt.Errorf("invalid kind <%v>", v)
				}
				kind, err := ParseResourceKind(k)
				if err != nil {
					return err
				}
				cfg.Kind = kind
				continue
			}

			// Any other config is added to the annotations
			cfg.Annotations[k] = v
		}

		// Deduplication will happen when upsertResource is called
		cfg.Refs = append(cfg.Refs, meta.Refs...)
	}

	switch cfg.Kind {
	case ResourceKindSource, ResourceKindModel:
		return p.parseSourceOrModelSQL(ctx, path, data, cfg)
	case ResourceKindMigration:
		return p.parseMigrationSQL(ctx, path, data, cfg)
	default:
		return fmt.Errorf("cannot use SQL for resource kind %q", cfg.Kind.String())
	}
}

// parseSourceOrModelSQL parses a SQL file for a source or model and upserts the resulting resource(s) to p.Resources.
// Sources and models defined in SQL are treated equally.
func (p *Parser) parseSourceOrModelSQL(ctx context.Context, path, data string, cfg sqlConfig) error {
	// If the connector is a DuckDB connector, enable DuckDB SQL-based inference.
	// Note: If the unspecified/default connector is DuckDB, duckDBConnectors will contain the empty string (see Parse).
	// (If templating was used, we skip DuckDB inference because the DuckDB parser may not be able to parse the templated code.)
	runDuckDBInference := false
	if !cfg.UsesTemplating {
		for _, c := range p.DuckDBConnectors {
			if c == cfg.Connector {
				runDuckDBInference = true
				break
			}
		}
	}

	// Extract info using DuckDB inference. DuckDB inference also supports rewriting embedded sources.
	embeddedSources := make(map[ResourceName]*runtimev1.SourceSpec)
	if runDuckDBInference {
		// Parse the SQL
		ast, err := duckdbsql.Parse(data)
		if err != nil {
			return fmt.Errorf("failed to parse DuckDB SQL: %w", newDuckDBError(err))
		}

		// Extract annotations into cfg
		annotations := ast.ExtractAnnotations()
		for _, a := range annotations {
			cfg.Annotations[a.Key] = a.Value
		}

		// Scan SQL for table references. Track references in refs and rewrite table functions into embedded sources.
		err = ast.RewriteTableRefs(func(t *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
			// Don't rewrite aliases
			if t.LocalAlias {
				return nil, false
			}

			// If embedded sources is enabled, parse it and add it to embeddedSources.
			if !cfg.DisableDuckDBSourceRewriting {
				name, spec, ok := parseEmbeddedSource(t, cfg.Connector)
				if ok {
					if embeddedSources[name] == nil {
						embeddedSources[name] = spec
						cfg.Refs = append(cfg.Refs, name)
					}
					return &duckdbsql.TableRef{Name: name.Name}, true
				}
			}

			// Not an embedded source. Add it to cfg.Refs if it's a regular table reference.
			if t.Name != "" && t.Function == "" && t.Path == "" {
				cfg.Refs = append(cfg.Refs, ResourceName{Name: t.Name})
			}
			return nil, false
		})
		if err != nil {
			return fmt.Errorf("error rewriting table refs: %w", err)
		}

		// Update data to the rewritten SQL
		sql, err := ast.Format()
		if err != nil {
			return fmt.Errorf("failed to format DuckDB SQL: %w", err)
		}
		data = sql
	}

	// Parse materialize from config
	var materialize *bool
	if v1, ok := cfg.Annotations["materialize"]; ok {
		switch v2 := v1.(type) {
		case bool:
			materialize = &v2
		case string:
			b, err := strconv.ParseBool(v2)
			if err != nil {
				return fmt.Errorf("invalid materialize value %q: %w", v2, err)
			}
			materialize = &b
		default:
			return fmt.Errorf("invalid materialize value <%v>", v1)
		}
	}

	// Parse timeout from config
	var timeoutSeconds int
	if v1, ok := cfg.Annotations["timeout"]; ok {
		d, err := parseDuration(v1)
		if err != nil {
			return err
		}
		timeoutSeconds = int(d.Seconds())
	}

	// Add the embedded sources
	for name, spec := range embeddedSources {
		r := p.upsertResource(ResourceKindSource, name.Name, path)

		// Since the same source may be referenced in multiple models with different TimeoutSeconds, we take the min of all the values.
		if spec.TimeoutSeconds < r.SourceSpec.TimeoutSeconds {
			spec.TimeoutSeconds = r.SourceSpec.TimeoutSeconds
		}

		// Since the embedded source's name is a hash of its parameters, we don't merge values into the existing spec.
		r.SourceSpec = spec
	}

	// Override the model
	r := p.upsertResource(ResourceKindModel, cfg.Name, path, cfg.Refs...)
	r.ModelSpec.Sql = strings.TrimSpace(data)
	r.ModelSpec.UsesTemplating = cfg.UsesTemplating
	if timeoutSeconds > 0 {
		r.ModelSpec.TimeoutSeconds = uint32(timeoutSeconds)
	}
	if cfg.Connector != "" {
		r.ModelSpec.Connector = cfg.Connector
	}
	if materialize != nil {
		r.ModelSpec.Materialize = materialize
	}
	if r.ModelSpec.Materialize == nil && cfg.Kind == ResourceKindSource {
		// If materialize was not set explicitly, always materialize sources
		b := true
		r.ModelSpec.Materialize = &b
	}

	return nil
}

// parseMigrationSQL parses a migration SQL file
func (p *Parser) parseMigrationSQL(ctx context.Context, path, data string, cfg sqlConfig) error {
	// Parse version from cfg
	var version uint
	if v, ok := cfg.Annotations["version"]; ok {
		switch v2 := v.(type) {
		case int:
			if v2 < 0 {
				return fmt.Errorf("invalid version value %d", v2)
			}
			version = uint(v2)
		case string:
			v3, err := strconv.ParseUint(v2, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid version value %q: %w", v2, err)
			}
			version = uint(v3)
		default:
			return fmt.Errorf("invalid version value <%v>", v)
		}
	}

	r := p.upsertResource(ResourceKindMigration, cfg.Name, path, cfg.Refs...)
	r.MigrationSpec.Sql = strings.TrimSpace(data)
	if cfg.Connector != "" {
		r.MigrationSpec.Connector = cfg.Connector
	}
	if version > 0 {
		r.MigrationSpec.Version = uint32(version)
	}

	return nil
}

// parseEmbeddedSource parses a table reference extracted from a DuckDB SQL query to a source spec.
// The returned name is derived from a hash of the source spec. It will be stable for any other table reference with equivalent path and properties.
func parseEmbeddedSource(t *duckdbsql.TableRef, sinkConnector string) (ResourceName, *runtimev1.SourceSpec, bool) {
	// The name can also potentially be a path
	path := t.Path
	if path == "" {
		path = t.Name
	}

	uri, err := url.Parse(path)
	if err != nil {
		return ResourceName{}, nil, false
	}

	// Applying some heuristics to determine if it's a path or just a table name.
	// If not a function and no protocol is in the path, we'll assume it's just a table name.
	if t.Function == "" && uri.Scheme == "" {
		return ResourceName{}, nil, false
	}

	if uri.Scheme == "" {
		uri.Scheme = "local_file"
	}

	_, ok := connectors.Connectors[uri.Scheme]
	if !ok {
		return ResourceName{}, nil, false
	}

	// TODO: Add support in DuckDB source for passing table function name directly, instead of "format".
	var format string
	switch t.Function {
	case "":
		format = ""
	case "read_parquet":
		format = "parquet"
	case "read_json", "read_json_auto", "read_ndjson", "read_ndjson_auto", "read_json_objects", "read_json_objects_auto", "read_ndjson_objects":
		format = "json"
	case "read_csv", "read_csv_auto":
		format = "csv"
	default:
		return ResourceName{}, nil, false
	}

	props := make(map[string]any)
	t.Properties["format"] = format
	t.Properties["path"] = uri.Path
	if t.Properties != nil {
		props["duckdb"] = t.Properties
	}

	propsPB, err := structpb.NewStruct(t.Properties)
	if err != nil {
		return ResourceName{}, nil, false
	}

	spec := &runtimev1.SourceSpec{}
	spec.SourceConnector = uri.Scheme
	spec.SinkConnector = sinkConnector
	spec.Properties = propsPB

	hash := md5.New()
	err = pbValueToHash(structpb.NewStructValue(propsPB), hash)
	if err != nil {
		return ResourceName{}, nil, false
	}
	_, err = hash.Write([]byte(spec.SourceConnector))
	if err != nil {
		return ResourceName{}, nil, false
	}
	_, err = hash.Write([]byte(spec.SinkConnector))
	if err != nil {
		return ResourceName{}, nil, false
	}

	name := ResourceName{Kind: ResourceKindSource, Name: "embed_" + hex.EncodeToString(hash.Sum(nil))}

	return name, spec, true
}

// pbValueToHash writes the contents of a structpb.Value to a hash writer in a deterministic order.
func pbValueToHash(v *structpb.Value, w io.Writer) error {
	switch v2 := v.Kind.(type) {
	case *structpb.Value_NullValue:
		_, err := w.Write([]byte{0})
		return err
	case *structpb.Value_NumberValue:
		err := binary.Write(w, binary.BigEndian, v2.NumberValue)
		return err
	case *structpb.Value_StringValue:
		_, err := w.Write([]byte(v2.StringValue))
		return err
	case *structpb.Value_BoolValue:
		err := binary.Write(w, binary.BigEndian, v2.BoolValue)
		return err
	case *structpb.Value_ListValue:
		for _, v3 := range v2.ListValue.Values {
			err := pbValueToHash(v3, w)
			if err != nil {
				return err
			}
		}
	case *structpb.Value_StructValue:
		// Iterate over sorted keys
		ks := maps.Keys(v2.StructValue.Fields)
		slices.Sort(ks)
		for _, k := range ks {
			_, err := w.Write([]byte(k))
			if err != nil {
				return err
			}
			err = pbValueToHash(v2.StructValue.Fields[k], w)
			if err != nil {
				return err
			}
		}
	default:
		panic(fmt.Sprintf("unknown kind %T", v.Kind))
	}
	return nil
}
