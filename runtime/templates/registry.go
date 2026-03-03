package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed definitions
var definitionsFS embed.FS

// Registry holds all loaded template definitions.
type Registry struct {
	templates map[string]*Template
	sorted    []*Template // sorted by name for stable List() output
}

// NewRegistry loads all embedded template definitions from the definitions/ directory tree.
func NewRegistry() (*Registry, error) {
	r := &Registry{
		templates: make(map[string]*Template),
	}

	err := fs.WalkDir(definitionsFS, "definitions", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		data, err := definitionsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		var t Template
		if err := yaml.Unmarshal(data, &t); err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		if t.Name == "" {
			return fmt.Errorf("template %s has no name", path)
		}

		if _, exists := r.templates[t.Name]; exists {
			return fmt.Errorf("duplicate template name %q in %s", t.Name, path)
		}

		r.templates[t.Name] = &t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading template definitions: %w", err)
	}

	// Build sorted list
	r.sorted = make([]*Template, 0, len(r.templates))
	for _, t := range r.templates {
		r.sorted = append(r.sorted, t)
	}
	sort.Slice(r.sorted, func(i, j int) bool {
		return r.sorted[i].Name < r.sorted[j].Name
	})

	return r, nil
}

// List returns all templates sorted by name.
func (r *Registry) List() []*Template {
	return r.sorted
}

// ListByTags returns templates that match ALL of the given tags.
// If tags is empty, returns all templates.
func (r *Registry) ListByTags(tags []string) []*Template {
	if len(tags) == 0 {
		return r.sorted
	}

	var result []*Template
	for _, t := range r.sorted {
		if matchesAllTags(t.Tags, tags) {
			result = append(result, t)
		}
	}
	return result
}

// Get returns a template by name, or nil and false if not found.
func (r *Registry) Get(name string) (*Template, bool) {
	t, ok := r.templates[name]
	return t, ok
}

// LookupByDriver finds the template for a given driver and output type.
// This is used by the backward-compatible GenerateTemplate RPC to map
// (driver, resource_type) pairs to template names.
func (r *Registry) LookupByDriver(driver, resourceType string) (*Template, bool) {
	switch resourceType {
	case "connector":
		// Connector templates are named after the driver
		return r.Get(driver)
	case "model":
		// Model templates: try driver-duckdb first (DuckDB rewrite), then driver-model
		if t, ok := r.Get(driver + "-duckdb"); ok {
			return t, true
		}
		if t, ok := r.Get(driver + "-model"); ok {
			return t, true
		}
		// Fallback: the driver itself might handle models (e.g. clickhouse-model)
		return nil, false
	}
	return nil, false
}

// matchesAllTags returns true if the template's tags contain all of the required tags.
func matchesAllTags(templateTags, requiredTags []string) bool {
	tagSet := make(map[string]bool, len(templateTags))
	for _, t := range templateTags {
		tagSet[t] = true
	}
	for _, req := range requiredTags {
		if !tagSet[req] {
			return false
		}
	}
	return true
}
