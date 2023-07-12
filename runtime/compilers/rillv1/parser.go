package rillv1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/slices"
)

// ErrInvalidProject indicates a project without a rill.yaml file
var ErrInvalidProject = errors.New("parser: not a valid project (rill.yaml not found)")

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

// ResourceName is a unique identifier for a resource
type ResourceName struct {
	Kind ResourceKind
	Name string
}

func (n ResourceName) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Name)
}

func (n ResourceName) Normalized() ResourceName {
	return ResourceName{
		Kind: n.Kind,
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

// upsertResource inserts or updates a resource in the parser's internal state.
// Upserting is required since both a YAML and SQL file may contribute information to the same resource.
// After calling upsertResource, the caller can modify the returned resource's spec, and should be cautious with overriding values that may have been set from another file.
func (p *Parser) upsertResource(kind ResourceKind, name, path string, refs ...ResourceName) *Resource {
	// Create the resource if not already present (ensures the spec for its kind is never nil)
	rn := ResourceName{Kind: kind, Name: name}
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
