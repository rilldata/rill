package rillv1

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

// sourceYAML is the raw structure of a Source resource defined in YAML (does not include common fields)
type sourceYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Type       string         `yaml:"type"`            // Backwards compatibility
	Timeout    string         `yaml:"timeout"`
	Refresh    *scheduleYAML  `yaml:"refresh"`
	Properties map[string]any `yaml:",inline" mapstructure:",remain"`
}

// parseSource parses a source definition and adds the resulting resource to p.Resources.
func (p *Parser) parseSource(ctx context.Context, node *Node) error {
	sqlProps := make(map[string]any)
	// If the source has SQL and hasn't specified a connector, we treat it as a model
	if node.SQL != "" && (node.Connector == "" || node.Connector == "duckdb") {
		var ok bool
		var err error
		sqlProps, ok, err = p.parseSQLSource(node)
		if err != nil {
			return err
		}
		// if cannot make this as sql source then treat as model
		if !ok {
			return p.parseModel(ctx, node)
		}
	}

	// Parse YAML
	tmp := &sourceYAML{}
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

	// Add SQL as a property
	if node.SQL != "" {
		if tmp.Properties == nil {
			tmp.Properties = map[string]any{}
		}
		tmp.Properties["sql"] = strings.TrimSpace(node.SQL)
	}
	// merge with sql sources props
	for k, v := range sqlProps {
		tmp.Properties[k] = v
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

	// Backward compatibility
	if tmp.Type != "" && node.Connector == "" {
		node.Connector = tmp.Type
	}

	// Validate the source has a connector
	if node.Connector == "" {
		return fmt.Errorf("must specify a connector")
	}

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}

	// Upsert source (in practice, this will always be an insert)
	// NOTE: After calling upsertResource, an error must not be returned. Any validation should be done before calling it.
	r := p.upsertResource(ResourceKindSource, node.Name, node.Paths, node.Refs...)
	r.SourceSpec.Properties = mergeStructPB(r.SourceSpec.Properties, props)
	if node.Connector != "" {
		r.SourceSpec.SourceConnector = node.Connector // Source connector. Sink connector not currently configurable.
	}
	if timeout != 0 {
		r.SourceSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	if schedule != nil {
		r.SourceSpec.RefreshSchedule = schedule
	}

	return nil
}

func (p *Parser) parseSQLSource(node *Node) (map[string]any, bool, error) {
	ast, err := duckdbsql.Parse(node.SQL)
	if err != nil {
		return nil, false, err
	}

	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return nil, false, nil
	}
	ref := refs[0]

	if len(ref.Paths) == 0 {
		return nil, false, nil
	}
	if len(ref.Paths) > 1 {
		return nil, false, nil
	}

	conn, ok := parseEmbeddedSourceConnector(ref.Paths[0], ref)
	if !ok {
		return nil, false, nil
	}

	props := make(map[string]any)

	switch conn {
	case "local_file":
		// TODO: get allow_root_access from env
		queryStr, err := rewriteLocalRelativePath(ast, p.Repo.Root(), false)
		if err != nil {
			return nil, false, err
		}
		node.Connector = "duckdb"
		props["sql"] = queryStr
	case "s3", "gcs":
		node.Connector = conn
		props["path"] = ref.Paths[0]
	default:
		return nil, false, nil
	}

	return props, true, nil
}

func rewriteLocalRelativePath(ast *duckdbsql.AST, repoRoot string, allowRootAccess bool) (string, error) {
	var resolveErr error
	err := ast.RewriteTableRefs(func(table *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
		newPaths := make([]string, 0)
		for _, p := range table.Paths {
			lp, err := fileutil.ResolveLocalPath(p, repoRoot, allowRootAccess)
			if err != nil {
				resolveErr = err
				return nil, false
			}
			newPaths = append(newPaths, lp)
		}

		return &duckdbsql.TableRef{
			Function:   table.Function,
			Paths:      newPaths,
			Properties: table.Properties,
		}, true
	})
	if resolveErr != nil {
		return "", resolveErr
	}
	if err != nil {
		return "", err
	}

	return ast.Format()
}

// scheduleYAML is the raw structure of a refresh schedule clause defined in YAML.
// This does not represent a stand-alone YAML file, just a partial used in other structs.
type scheduleYAML struct {
	Cron   string `yaml:"cron" mapstructure:"cron"`
	Ticker string `yaml:"ticker" mapstructure:"ticker"`
}

func parseScheduleYAML(raw *scheduleYAML) (*runtimev1.Schedule, error) {
	if raw == nil || (raw.Cron == "" && raw.Ticker == "") {
		return nil, nil
	}

	s := &runtimev1.Schedule{}
	if raw.Cron != "" {
		_, err := cron.ParseStandard(raw.Cron)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		s.Cron = raw.Cron
	}

	if raw.Ticker != "" {
		d, err := parseDuration(raw.Ticker)
		if err != nil {
			return nil, fmt.Errorf("invalid ticker: %w", err)
		}
		s.TickerSeconds = uint32(d.Seconds())
	}

	return s, nil
}

// parseDuration parses a value into a time duration.
// If no unit is specified, it assumes the value is in seconds.
func parseDuration(v any) (time.Duration, error) {
	switch v := v.(type) {
	case int:
		return time.Duration(v) * time.Second, nil
	case string:
		// Try parsing as an int first
		res, err := strconv.Atoi(v)
		if err == nil {
			return time.Duration(res) * time.Second, nil
		}
		// Try parsing with a unit
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("invalid time duration value %v: %w", v, err)
		}
		return d, nil
	default:
		return 0, fmt.Errorf("invalid time duration value <%v>", v)
	}
}

// mergeStructPB merges two structpb.Structs, with b taking precedence over a.
// It overwrites a in place and returns it.
func mergeStructPB(a, b *structpb.Struct) *structpb.Struct {
	if a == nil || a.Fields == nil {
		return b
	}
	if b == nil || b.Fields == nil {
		return a
	}
	for k, v := range b.Fields {
		a.Fields[k] = v
	}
	return a
}
