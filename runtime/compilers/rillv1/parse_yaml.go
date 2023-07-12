package rillv1

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	Title       string
	Description string
	Connectors  []*ConnectorDef
	Variables   []*VariableDef
}

// ConnectorDef is a subtype of RillYAML, defining connectors required by the project
type ConnectorDef struct {
	Type     string
	Name     string
	Defaults map[string]string
}

// VariableDef is a subtype of RillYAML, defining defaults for project variables
type VariableDef struct {
	Name    string
	Default string
}

// rillYAML is the raw YAML structure of rill.yaml
type rillYAML struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env"`
	Connectors  []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
}

// parseRillYAML parses rill.yaml
func (p *Parser) parseRillYAML(ctx context.Context, data string) error {
	tmp := &rillYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return fmt.Errorf("rill.yaml: %w", err)
	}

	res := &RillYAML{
		Title:       tmp.Title,
		Description: tmp.Description,
		Connectors:  make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:   make([]*VariableDef, len(tmp.Env)),
	}

	for i, c := range tmp.Connectors {
		res.Connectors[i] = &ConnectorDef{
			Type:     c.Type,
			Name:     c.Name,
			Defaults: c.Defaults,
		}
	}

	i := 0
	for k, v := range tmp.Env {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}

// genericYAML contains common fields that any YAML file in a Rill project can specify.
type genericYAML struct {
	// Kind can be inferred from the directory name in certain cases, but otherwise must be specified manually.
	Kind *string `yaml:"kind"`
	// Name is usually inferred from the filename, but can be specified manually.
	Name string `yaml:"name"`
	// Refs are a list of other resources that this resource depends on. They are usually inferred from other fields, but can also be specified manually.
	Refs []*yaml.Node `yaml:"refs"`
}

// parseYAML parses a YAML file and adds the resulting resource(s) to p.Resources.
func (p *Parser) parseYAML(ctx context.Context, path, data string) error {
	// We treat the "sources", "models", and "dashboards" directories as providing special context.
	// Files outside must specify a "kind" in the YAML.
	var kind ResourceKind
	if strings.HasPrefix(path, "/sources") {
		kind = ResourceKindSource
	} else if strings.HasPrefix(path, "/models") {
		kind = ResourceKindModel
	} else if strings.HasPrefix(path, "/dashboards") {
		kind = ResourceKindMetricsView
	} else {
		tmp := &genericYAML{}
		if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
			return fmt.Errorf("YAML error: %w", err)
		}
		if tmp.Kind == nil {
			// If no Kind is specified, we assume the file is not a Rill resource
			return nil
		}
		var err error
		kind, err = ParseResourceKind(*tmp.Kind)
		if err != nil {
			return err
		}
	}

	switch kind {
	case ResourceKindSource:
		return p.parseSourceYAML(ctx, path, data)
	case ResourceKindModel:
		return p.parseModelYAML(ctx, path, data)
	case ResourceKindMetricsView:
		return p.parseMetricsViewYAML(ctx, path, data)
	case ResourceKindMigration:
		return p.parseMigrationYAML(ctx, path, data)
	default:
		panic(fmt.Errorf("unexpected resource kind: %s", kind.String()))
	}
}

// sourceYAML is the raw structure of a Source resource defined in YAML
type sourceYAML struct {
	genericYAML `yaml:",inline"`
	Connector   string         `yaml:"connector"` // Source connector. Sink connector not currently supported.
	Type        string         `yaml:"type"`      // Backwards compatibility
	Timeout     *string        `yaml:"timeout"`
	Refresh     *scheduleYAML  `yaml:"refresh"`
	Properties  map[string]any `yaml:",inline"`
}

// parseModelYAML parses a source YAML file and adds the resulting resource to p.Resources.
func (p *Parser) parseSourceYAML(ctx context.Context, path, data string) error {
	// Parse the YAML and handle generic fields
	tmp := &sourceYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return fmt.Errorf("YAML error: %w", err)
	}
	if tmp.Name == "" {
		tmp.Name = fileutil.Stem(path)
	}
	refs, err := parseYAMLRefs(tmp.Refs)
	if err != nil {
		return err
	}

	// Backward compatibility
	if tmp.Type != "" && tmp.Connector == "" {
		tmp.Connector = tmp.Type
	}

	timeout, err := parseDuration(tmp.Timeout)
	if err != nil {
		return err
	}

	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}

	r := p.upsertResource(ResourceKindSource, tmp.Name, path, refs...)
	r.SourceSpec.SourceConnector = tmp.Connector
	r.SourceSpec.Properties = props
	r.SourceSpec.TimeoutSeconds = uint32(timeout.Seconds())
	r.SourceSpec.RefreshSchedule = schedule

	return nil
}

// modelYAML is the raw structure of a Model resource defined in YAML
type modelYAML struct {
	genericYAML `yaml:",inline"`
	Connector   string        `yaml:"connector"`
	Materialize *bool         `yaml:"materialize"`
	Timeout     *string       `yaml:"timeout"`
	Refresh     *scheduleYAML `yaml:"refresh"`
}

// parseModelYAML parses a model YAML file and adds the resulting resource to p.Resources.
func (p *Parser) parseModelYAML(ctx context.Context, path, data string) error {
	tmp := &modelYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return fmt.Errorf("YAML error: %w", err)
	}
	if tmp.Name == "" {
		tmp.Name = fileutil.Stem(path)
	}
	refs, err := parseYAMLRefs(tmp.Refs)
	if err != nil {
		return err
	}

	timeout, err := parseDuration(tmp.Timeout)
	if err != nil {
		return err
	}

	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	r := p.upsertResource(ResourceKindModel, tmp.Name, path, refs...)
	r.ModelSpec.Connector = tmp.Connector
	r.ModelSpec.Materialize = tmp.Materialize
	r.ModelSpec.TimeoutSeconds = uint32(timeout.Seconds())
	r.ModelSpec.RefreshSchedule = schedule

	return nil
}

// metricsViewYAML is the raw structure of a MetricsView resource defined in YAML
type metricsViewYAML struct {
	genericYAML        `yaml:",inline"`
	Title              string   `yaml:"title"`
	DisplayName        string   `yaml:"display_name"` // Backwards compatibility
	Description        string   `yaml:"description"`
	Model              string   `yaml:"model"`
	TimeDimension      string   `yaml:"timeseries"`
	SmallestTimeGrain  string   `yaml:"smallest_time_grain"`
	DefaultTimeRange   string   `yaml:"default_time_range"`
	AvailableTimeZones []string `yaml:"available_time_zones"`
	Dimensions         []*struct {
		Name        string
		Label       string
		Column      string
		Property    string // For backwards compatibility
		Description string
		Ignore      bool `yaml:"ignore"`
	}
	Measures []*struct {
		Name                string
		Label               string
		Expression          string
		Description         string
		Format              string `yaml:"format_preset"`
		Ignore              bool   `yaml:"ignore"`
		ValidPercentOfTotal bool   `yaml:"valid_percent_of_total"`
	}
}

// parseMetricsViewYAML parses a metrics view (dashboard) YAML file and adds the resulting resource to p.Resources.
func (p *Parser) parseMetricsViewYAML(ctx context.Context, path, data string) error {
	// Parse the YAML and handle generic fields
	tmp := &metricsViewYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return fmt.Errorf("YAML error: %w", err)
	}
	if tmp.Name == "" {
		tmp.Name = fileutil.Stem(path)
	}
	refs, err := parseYAMLRefs(tmp.Refs)
	if err != nil {
		return err
	}

	// Backwards compatibility
	if tmp.DisplayName != "" && tmp.Title == "" {
		tmp.Title = tmp.DisplayName
	}

	smallestTimeGrain, err := parseTimeGrain(tmp.SmallestTimeGrain)
	if err != nil {
		return fmt.Errorf(`invalid "smallest_time_grain": %w`, err)
	}

	if tmp.DefaultTimeRange != "" {
		_, err := duration.ParseISO8601(tmp.DefaultTimeRange)
		if err != nil {
			return fmt.Errorf(`invalid "default_time_range": %w`, err)
		}
	}

	for _, tz := range tmp.AvailableTimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}

	r := p.upsertResource(ResourceKindModel, tmp.Name, path, refs...)
	spec := r.MetricsViewSpec

	spec.Title = tmp.Title
	spec.Description = tmp.Description
	spec.Model = tmp.Model
	spec.TimeDimension = tmp.TimeDimension
	spec.SmallestTimeGrain = smallestTimeGrain
	spec.DefaultTimeRange = tmp.DefaultTimeRange
	spec.AvailableTimeZones = tmp.AvailableTimeZones

	for i, dim := range tmp.Dimensions {
		if dim.Ignore {
			continue
		}

		// Backwards compatibility
		if dim.Property != "" && dim.Column == "" {
			dim.Column = dim.Property
		}

		// Backwards compatibility
		if dim.Name == "" {
			if dim.Column == "" {
				dim.Name = fmt.Sprintf("dimension_%d", i)
			} else {
				dim.Name = dim.Column
			}
		}

		spec.Dimensions = append(spec.Dimensions, &runtimev1.MetricsViewSpec_Dimension{
			Name:        dim.Name,
			Column:      dim.Column,
			Label:       dim.Label,
			Description: dim.Description,
		})
	}

	for i, measure := range tmp.Measures {
		if measure.Ignore {
			continue
		}

		// Backwards compatibility
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}

		spec.Measures = append(spec.Measures, &runtimev1.MetricsViewSpec_Measure{
			Name:                measure.Name,
			Expression:          measure.Expression,
			Label:               measure.Label,
			Description:         measure.Description,
			Format:              measure.Format,
			ValidPercentOfTotal: measure.ValidPercentOfTotal,
		})
	}

	return nil
}

// migrationYAML is the raw structure of a Migration resource defined in YAML
type migrationYAML struct {
	genericYAML `yaml:",inline"`
	Connector   string `yaml:"connector"`
	Version     uint   `yaml:"version"`
	SQL         string `yaml:"sql"`
}

// parseMigrationYAML parses a migration YAML file and adds the resulting resource to p.Resources.
func (p *Parser) parseMigrationYAML(ctx context.Context, path, data string) error {
	tmp := &migrationYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return fmt.Errorf("YAML error: %w", err)
	}
	if tmp.Name == "" {
		tmp.Name = fileutil.Stem(path)
	}
	refs, err := parseYAMLRefs(tmp.Refs)
	if err != nil {
		return err
	}

	r := p.upsertResource(ResourceKindMigration, tmp.Name, path, refs...)
	r.MigrationSpec.Connector = tmp.Connector
	r.MigrationSpec.Version = uint32(tmp.Version)
	r.MigrationSpec.Sql = tmp.SQL

	return nil
}

// parseYAMLRefs parses a list of YAML nodes into a list of ResourceNames.
// It's used to parse the "refs" field in genericYAML.
func parseYAMLRefs(refs []*yaml.Node) ([]ResourceName, error) {
	var res []ResourceName
	for _, ref := range refs {
		// We support string refs of the form "my-resource" and "Kind/my-resource"
		if ref.Kind == yaml.ScalarNode {
			var identifier string
			err := ref.Decode(&identifier)
			if err != nil {
				return nil, fmt.Errorf("invalid refs: %v", ref)
			}

			// Parse name and kind from identifier
			parts := strings.Split(identifier, "/")
			if len(parts) != 1 && len(parts) != 2 {
				return nil, fmt.Errorf("invalid refs: invalid identifier %q", identifier)
			}

			var name ResourceName
			if len(parts) == 1 {
				name.Name = parts[0]
			} else {
				// Kind and name specified
				kind, err := ParseResourceKind(parts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid refs: %w", err)
				}
				name.Kind = kind
				name.Name = parts[1]
			}
			res = append(res, name)
			continue
		}

		// We support map refs of the form { kind: "kind", name: "my-resource" }
		if ref.Kind == yaml.MappingNode {
			var name ResourceName
			err := ref.Decode(&name)
			if err != nil {
				return nil, fmt.Errorf("invalid refs: %w", err)
			}
			res = append(res, name)
			continue
		}

		// ref was neither a string nor a map
		return nil, fmt.Errorf("invalid refs: %v", ref)
	}
	return res, nil
}

// scheduleYAML is the raw structure of a refresh schedule clause defined in YAML.
// This does not represent a stand-alone YAML file, just a partial used in other structs.
type scheduleYAML struct {
	Cron   string `yaml:"cron"`
	Ticker string `yaml:"ticker"`
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
	if v == nil {
		return 0, nil
	}
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

// parseTimeGrain parses a YAML time grain string
func parseTimeGrain(s string) (runtimev1.TimeGrain, error) {
	switch strings.ToLower(s) {
	case "":
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, nil
	case "ms", "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, nil
	case "s", "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND, nil
	case "min", "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE, nil
	case "h", "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR, nil
	case "d", "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY, nil
	case "w", "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK, nil
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH, nil
	case "q", "quarter":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER, nil
	case "y", "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR, nil
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, fmt.Errorf("invalid time grain %q", s)
	}
}
