package rillv1

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// ErrInvalidProject indicates a project without a rill.yaml file
var ErrInvalidProject = errors.New("parser: not a valid project (rill.yaml not found)")

// ResourceName is a unique identifier for a resource
type ResourceName struct {
	Kind string
	Name string
}

func (n ResourceName) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Name)
}

func (n ResourceName) Normalized() ResourceName {
	return ResourceName{
		Kind: strings.ToLower(n.Kind),
		Name: strings.ToLower(n.Name),
	}
}

// ResourceKind identifies a resource type supported by the parser
type ResourceKind int

const (
	ResourceKindUnspecified ResourceKind = iota
	ResourceKindSource
	ResourceKindModel
	ResourceKindMetricsView
	ResourceKindMigration
)

func ParseResourceKind(kind string) (ResourceKind, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "":
		return ResourceKindUnspecified, nil
	case "source":
		return ResourceKindSource, nil
	case "model":
		return ResourceKindModel, nil
	case "metricsview", "metrics_view", "dashboard":
		return ResourceKindMetricsView, nil
	case "migration":
		return ResourceKindMigration, nil
	default:
		return ResourceKindUnspecified, fmt.Errorf("invalid resource kind %q", kind)
	}
}

func (k ResourceKind) String() string {
	switch k {
	case ResourceKindUnspecified:
		return ""
	case ResourceKindSource:
		return "Source"
	case ResourceKindModel:
		return "Model"
	case ResourceKindMetricsView:
		return "MetricsView"
	case ResourceKindMigration:
		return "Migration"
	default:
		panic(fmt.Sprintf("unexpected resource kind: %d", k))
	}
}

// Resource parsed from code files.
// One file may output multiple resources and multiple files may contribute config to one resource.
type Resource struct {
	// Metadata
	Name  ResourceName
	Refs  []ResourceName
	Paths []string

	// Only one of these will be non-nil
	SourceSpec      *runtimev1.SourceSpec
	ModelSpec       *runtimev1.ModelSpec
	MetricsViewSpec *runtimev1.MetricsViewSpec
	MigrationSpec   *runtimev1.MigrationSpec
}

// Diff shows changes to Parser.Resources following an incremental reparse.
type Diff struct {
	Added    []ResourceName
	Modified []ResourceName
	Deleted  []ResourceName
}

// Parser parses a Rill project directory into a set of resources.
// After the initial parse, the parser can be used to incrementally reparse a subset of files.
// Parser is not concurrency safe.
type Parser struct {
	// Options
	Repo             drivers.RepoStore
	InstanceID       string
	DuckDBConnectors []string

	// Output
	RillYAML  *RillYAML
	Resources map[ResourceName]*Resource
	Errors    []*runtimev1.ParseError

	// Internal state
	resourcesForPath  map[string][]*Resource
	insertedResources []*Resource
}

// ParseRillYAML parses only the project's rill.yaml (or rill.yml) file.
func ParseRillYAML(ctx context.Context, repo drivers.RepoStore, instanceID string) (*RillYAML, error) {
	paths, err := repo.ListRecursive(ctx, instanceID, "rill.{yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("could not list project files: %w", err)
	}

	p := Parser{Repo: repo, InstanceID: instanceID}
	err = p.parsePaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	return p.RillYAML, nil
}

// Parse creates a new parser and parses the entire project.
//
// Note on SQL parsing: For DuckDB SQL specifically, the parser can use a SQL parser to extract refs and annotations (instead of relying on templating or YAML).
// To enable SQL parsing for a connector, pass it in duckDBConnectors. If DuckDB SQL parsing should be used on files where no connector is specified, put an empty string in duckDBConnectors.
func Parse(ctx context.Context, repo drivers.RepoStore, instanceID string, duckDBConnectors []string) (*Parser, error) {
	p := &Parser{
		Repo:             repo,
		InstanceID:       instanceID,
		DuckDBConnectors: duckDBConnectors,
		Resources:        make(map[ResourceName]*Resource),
		resourcesForPath: make(map[string][]*Resource),
	}

	paths, err := p.Repo.ListRecursive(ctx, p.InstanceID, "**/*.{sql,yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("could not list project files: %w", err)
	}

	err = p.parsePaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Reparse re-parses the indicated file paths, updating the Parser's state.
// If a previous call to Reparse has returned an error, the Parser may not be accessed or called again.
func (p *Parser) Reparse(ctx context.Context, paths []string) (*Diff, error) {
	// The logic here is slightly tricky because the relationship between files and resources can vary:
	//
	// - Case 1: one file created one resource
	// - Case 2: one file created multiple resources
	// - Case 3: multiple files contributed to one resource (for example, "model.sql" and "model.yaml")
	//
	// The high-level approach is: We'll delete all existing resources *related* to the changed paths and (re)parse them.
	// Then at the end, we build a diff that treats any resource that was both "deleted" and "added" as an "update".
	// (Renames are not supported in the parser. It needs to be handled by the caller, since parser state alone is insufficient to detect it.)

	// Phase 1: Clear existing state related to the paths.
	// Identify all paths directly passed and paths indirectly related through resourcesForPath and Resource.Paths.
	// And delete all resources and parse errors related to those paths.
	var parsePaths []string            // Paths we should pass to parsePaths
	var deletedResources []*Resource   // Resources deleted in Phase 1 (some may be added back in Phase 2)
	checkPaths := paths                // Paths we should visit in the loop
	seenPaths := make(map[string]bool) // Paths already visited by the loop
	for i := 0; i < len(checkPaths); i++ {
		// Don't check the same path twice
		path := checkPaths[i]
		if seenPaths[path] {
			continue
		}
		seenPaths[path] = true

		// Skip files that aren't SQL or YAML
		if !strings.HasSuffix(path, ".sql") && !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			continue
		}

		// If a file exists at path, add it to the parse list
		_, err := p.Repo.Stat(ctx, p.InstanceID, path)
		if err == nil {
			parsePaths = append(parsePaths, path)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("unexpected file stat error: %w", err)
		}

		// Check if path is rill.yaml and clear it (so we can re-parse it)
		if path == "/rill.yaml" || path == "/rill.yml" {
			p.RillYAML = nil
		}

		// Remove all resources derived from this path, and add any related paths to the check list
		for _, resource := range p.resourcesForPath[path] {
			// Multiple entries in resourcesForPath may point to the same resource.
			// By adding resource.Paths to checkPaths, the outer loop will eventually clear those (maybe it already has).
			checkPaths = append(checkPaths, resource.Paths...)
			// But make sure we only track each deleted resource once
			if _, ok := p.Resources[resource.Name.Normalized()]; ok {
				delete(p.Resources, resource.Name.Normalized())
				deletedResources = append(deletedResources, resource)
			}
		}
		delete(p.resourcesForPath, path)

		// Remove all parse errors related to this path
		// (We can't mutate p.Errors while iterating over it, hence the nested loop here.)
		for {
			found := false
			for i, err := range p.Errors {
				if err.FilePath == path {
					found = true
					p.Errors = slices.Delete(p.Errors, i, i+1)
					break
				}
			}
			if !found {
				break
			}
		}
	}

	// Phase 2: Parse (or reparse) the related paths, adding back resources
	err := p.parsePaths(ctx, parsePaths)
	if err != nil {
		return nil, err
	}

	// Phase 3: Build the diff using p.insertedResources and deletedResources
	diff := &Diff{}
	for _, resource := range p.insertedResources {
		addedBack := false
		for _, deleted := range deletedResources {
			if resource.Name.Normalized() == deleted.Name.Normalized() {
				addedBack = true
				break
			}
		}
		if addedBack {
			diff.Modified = append(diff.Modified, resource.Name)
		} else {
			diff.Added = append(diff.Added, resource.Name)
		}
	}
	for _, deleted := range deletedResources {
		if p.Resources[deleted.Name.Normalized()] == nil {
			diff.Deleted = append(diff.Deleted, deleted.Name)
		}
	}

	return diff, nil
}

// parsePaths is the internal entrypoint for parsing a list of paths.
// It assumes that the caller has already checked that the paths exist.
// It also assumes that the caller has already removed any previous resources related to the paths,
// enabling parsePaths to upsert changes, enabling multiple files to provide data for one resource
// (like "my-model.sql" and "my-model.yaml").
func (p *Parser) parsePaths(ctx context.Context, paths []string) error {
	// Reset insertedResources on each parse (only used to construct Diff in Reparse)
	p.insertedResources = nil

	// Sort paths such that we parse YAML files before SQL files.
	// This enables YAML parsers to assign properties to specs without checking for prior existence.
	// It also enables p.parseSQL to obtain context provided in a YAML file, e.g. to guide template and/or DuckDB SQL parsing.
	slices.SortFunc(paths, func(a, b string) bool {
		return strings.HasSuffix(b, ".sql")
	})

	// Parse and upsert each path
	for _, path := range paths {
		// Handle SQL and YAML files, skip any other file
		var isSQL, isYAML bool
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			isYAML = true
		} else if strings.HasSuffix(path, ".sql") {
			isSQL = true
		} else {
			continue
		}

		// Load contents
		data, err := p.Repo.Get(ctx, p.InstanceID, path)
		if err != nil {
			// TODO: Handle dirty parses where files disappear during parsing
			return err
		}

		// Handle rill.yaml separately (if parsing of rill.yaml fails, we exit early instead of adding a ParseError)
		if path == "/rill.yaml" || path == "/rill.yml" {
			err := p.parseRillYAML(ctx, data)
			if err != nil {
				return err
			}
			continue
		}

		// Parse file, accumulating errors in p.Errors
		if isYAML {
			err = p.parseYAML(ctx, path, data)
		} else if isSQL {
			err = p.parseSQL(ctx, path, data)
		}
		if err != nil {
			p.Errors = append(p.Errors, &runtimev1.ParseError{
				Message:  err.Error(),
				FilePath: path,
			})
		}
	}

	// If we didn't encounter rill.yaml, that's a breaking error
	if p.RillYAML == nil {
		return ErrInvalidProject
	}

	return nil
}

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
	Timeout     int32          `yaml:"timeout"`
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

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}

	r := p.upsertResource(ResourceKindSource, tmp.Name, path, refs...)
	r.SourceSpec.SourceConnector = tmp.Connector
	r.SourceSpec.Properties = props
	r.SourceSpec.TimeoutSeconds = uint32(tmp.Timeout)

	return nil
}

// modelYAML is the raw structure of a Model resource defined in YAML
type modelYAML struct {
	genericYAML `yaml:",inline"`
	Connector   string `yaml:"connector"`
	Materialize *bool  `yaml:"materialize"`
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

	r := p.upsertResource(ResourceKindModel, tmp.Name, path, refs...)
	r.ModelSpec.Connector = tmp.Connector
	r.ModelSpec.Materialize = tmp.Materialize

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
			return fmt.Errorf("error parsing companion YAML file at %q: %w", yamlPath, err)
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
		// Extract annotations into cfg
		annotations, err := duckdbsql.ExtractAnnotations(data)
		if err != nil {
			return fmt.Errorf("error extracting annotations: %w", err)
		}
		for _, a := range annotations {
			cfg.Annotations[a.Key] = a.Value
		}

		// Scan SQL for table references. Track references in refs and rewrite table functions into embedded sources.
		sql, err := duckdbsql.RewriteTableRefs(data, func(t *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
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
			// TODO: Should we have a better way of ensuring that t.Name is not an alias?
			if t.Name != "" && t.Function == "" && t.Path == "" {
				cfg.Refs = append(cfg.Refs, ResourceName{Kind: "", Name: t.Name})
			}
			return nil, false
		})
		if err != nil {
			return fmt.Errorf("error rewriting table refs: %w", err)
		}

		// Update data to the rewritten SQL
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
		switch v2 := v1.(type) {
		case int:
			timeoutSeconds = v2
		case string:
			d, err := time.ParseDuration(v2)
			if err != nil {
				return fmt.Errorf("invalid timeout value %q: %w", v2, err)
			}
			timeoutSeconds = int(d.Seconds())
		default:
			return fmt.Errorf("invalid timeout value <%v>", v1)
		}
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
	r.ModelSpec.Sql = data
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
	r.MigrationSpec.Sql = data
	if cfg.Connector != "" {
		r.MigrationSpec.Connector = cfg.Connector
	}
	if version > 0 {
		r.MigrationSpec.Version = uint32(version)
	}

	return nil
}

// upsertResource inserts or updates a resource in the parser's internal state.
// Upserting is required since both a YAML and SQL file may contribute information to the same resource.
// After calling upsertResource, the caller can modify the returned resource's spec, and should be cautious with overriding values that may have been set from another file.
func (p *Parser) upsertResource(kind ResourceKind, name, path string, refs ...ResourceName) *Resource {
	// Create the resource if not already present (ensures the spec for its kind is never nil)
	rn := ResourceName{Kind: kind.String(), Name: name}
	r, ok := p.Resources[rn.Normalized()]
	if !ok {
		r = &Resource{Name: rn}
		p.Resources[rn.Normalized()] = r
		p.insertedResources = append(p.insertedResources, r)
		switch kind {
		case ResourceKindSource:
			r.SourceSpec = &runtimev1.SourceSpec{}
		case ResourceKindModel:
			r.ModelSpec = &runtimev1.ModelSpec{}
		case ResourceKindMetricsView:
			r.MetricsViewSpec = &runtimev1.MetricsViewSpec{}
		case ResourceKindMigration:
			r.MigrationSpec = &runtimev1.MigrationSpec{}
		default:
			panic(fmt.Errorf("unexpected resource kind: %s", kind.String()))
		}
	}

	// Index path if not already present
	found := false
	for _, p := range r.Paths {
		if p == path {
			found = true
			break
		}
	}
	if !found {
		r.Paths = append(r.Paths, path)
		p.resourcesForPath[path] = append(p.resourcesForPath[path], r)
	}

	// Add refs that are not already present
	for _, refA := range refs {
		found := false
		for _, refB := range r.Refs {
			if refA.Normalized() == refB.Normalized() {
				found = true
				break
			}
		}
		if !found {
			r.Refs = append(r.Refs, refA)
		}
	}

	return r
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
				name.Kind = kind.String()
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

// parseEmbeddedSource parses a table reference extracted from a DuckDB SQL query to a source spec.
// The returned name is derived from a hash of the source spec. It will be stable for any other table reference with equivalent path and properties.
func parseEmbeddedSource(t *duckdbsql.TableRef, sinkConnector string) (ResourceName, *runtimev1.SourceSpec, bool) {
	if t.Path == "" {
		return ResourceName{}, nil, false
	}

	uri, err := url.Parse(t.Path)
	if err != nil {
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

	name := ResourceName{Kind: ResourceKindSource.String(), Name: "embed_" + hex.EncodeToString(hash.Sum(nil))}

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
