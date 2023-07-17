package rillv1

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
)

// modelYAML is the raw structure of a Model resource defined in YAML (does not include common fields)
type modelYAML struct {
	Materialize  *bool         `yaml:"materialize" mapstructure:"materialize"`
	Timeout      string        `yaml:"timeout" mapstructure:"timeout"`
	Refresh      *scheduleYAML `yaml:"refresh" mapstructure:"refresh"`
	ParserConfig struct {
		DuckDB struct {
			InferRefs      *bool `yaml:"infer_refs" mapstructure:"infer_refs"`
			RewriteSources *bool `yaml:"rewrite_sources" mapstructure:"rewrite_sources"`
		} `yaml:"duckdb" mapstructure:"duckdb"`
	} `yaml:"parser"`
}

// parseModel parses a model definition and adds the resulting resource to p.Resources.
func (p *Parser) parseModel(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &modelYAML{}
	if node.YAML != nil {
		if err := node.YAML.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Override YAML config with SQL annotations
	err := mapstructureUnmarshal(node.SQLAnnotations, tmp)
	if err != nil {
		return pathError{path: node.SQLPath, err: fmt.Errorf("invalid SQL annotations: %w", err)}
	}

	// Parse timeout
	var timeout time.Duration
	if tmp.Timeout != "" {
		timeout, err = parseDuration(tmp.Timeout)
		if err != nil {
			return err
		}
	}

	// Parse refresh schedule
	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	// If the connector is a DuckDB connector, extract info using DuckDB SQL parsing. This also supports rewriting embedded sources.
	// (If templating was used, we skip DuckDB inference because the DuckDB parser may not be able to parse the templated code.)
	isDuckDB := false
	for _, c := range p.DuckDBConnectors {
		// Note: If the unspecified/default connector is DuckDB, duckDBConnectors will contain the empty string (see Parse).
		if c == node.Connector {
			isDuckDB = true
			break
		}
	}
	duckDBInferRefs := true
	if tmp.ParserConfig.DuckDB.InferRefs != nil {
		duckDBInferRefs = *tmp.ParserConfig.DuckDB.InferRefs
	}
	duckDBRewriteSources := true
	if tmp.ParserConfig.DuckDB.RewriteSources != nil {
		duckDBRewriteSources = *tmp.ParserConfig.DuckDB.RewriteSources
	}
	refs := node.Refs
	embeddedSources := make(map[ResourceName]*runtimev1.SourceSpec)
	if isDuckDB && !node.SQLUsesTemplating && node.SQL != "" && (duckDBInferRefs || duckDBRewriteSources) {
		// Parse the SQL
		ast, err := duckdbsql.Parse(node.SQL)
		if err != nil {
			return pathError{path: node.SQLPath, err: newDuckDBError(err)}
		}

		// Scan SQL for table references. Track references in refs and rewrite table functions into embedded sources.
		err = ast.RewriteTableRefs(func(t *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
			// Don't rewrite aliases
			if t.LocalAlias {
				return nil, false
			}

			// If embedded sources is enabled, parse it and add it to embeddedSources.
			if duckDBRewriteSources {
				name, spec, ok := parseEmbeddedSource(t, node.Connector)
				if ok {
					if embeddedSources[name] == nil {
						spec.TimeoutSeconds = uint32(timeout.Seconds())
						embeddedSources[name] = spec
						refs = append(refs, name)
					}
					return &duckdbsql.TableRef{Name: name.Name}, true
				}
			}

			// Not an embedded source. Add it to refs if it's a regular table reference.
			if duckDBInferRefs {
				if t.Name != "" && t.Function == "" && t.Path == "" {
					refs = append(refs, ResourceName{Name: t.Name})
				}
			}
			return nil, false
		})
		if err != nil {
			return pathError{path: node.SQLPath, err: fmt.Errorf("error rewriting table refs: %w", err)}
		}

		// Update data to the rewritten SQL
		sql, err := ast.Format()
		if err != nil {
			return pathError{path: node.SQLPath, err: fmt.Errorf("failed to format DuckDB SQL: %w", err)}
		}
		node.SQL = sql
	}

	// Add the embedded sources
	for name, spec := range embeddedSources {
		r := p.upsertResource(ResourceKindSource, name.Name, node.Paths)

		// Since the same source may be referenced in multiple models with different TimeoutSeconds, coerce it to the lowest timeout of any referencing model.
		if spec.TimeoutSeconds == 0 || spec.TimeoutSeconds > r.SourceSpec.TimeoutSeconds {
			spec.TimeoutSeconds = r.SourceSpec.TimeoutSeconds
		}

		// Since the embedded source's name is a hash of its parameters, we don't merge values into the existing spec.
		r.SourceSpec = spec
	}

	// Upsert the model
	r := p.upsertResource(ResourceKindModel, node.Name, node.Paths, refs...)
	if node.SQL != "" {
		r.ModelSpec.Sql = strings.TrimSpace(node.SQL)
		r.ModelSpec.UsesTemplating = node.SQLUsesTemplating
	}
	if node.Connector != "" {
		r.ModelSpec.Connector = node.Connector
	}
	if tmp.Materialize != nil {
		r.ModelSpec.Materialize = tmp.Materialize
	}
	if timeout > 0 {
		r.ModelSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	if schedule != nil {
		r.ModelSpec.RefreshSchedule = schedule
	}

	// parseSource calls parseModel for SQL sources without a connector. Materialize such models if they don't use embedded sources.
	if node.Kind == ResourceKindSource && r.ModelSpec.Materialize == nil && len(embeddedSources) == 0 {
		b := true
		r.ModelSpec.Materialize = &b
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

	// NOTE: Using url.Parse is a little hacky. The first path component will be parsed as the host (so don't rely on uri.Path!)
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

	_, ok := drivers.Connectors[uri.Scheme]
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
	props["path"] = path
	if format != "" {
		props["format"] = format
	}
	if t.Properties != nil {
		props["duckdb"] = t.Properties
	}

	propsPB, err := structpb.NewStruct(props)
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
