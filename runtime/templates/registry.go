package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
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

	err := fs.WalkDir(definitionsFS, "definitions", func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		data, err := definitionsFS.ReadFile(fpath)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", fpath, err)
		}

		// Skip stub files (placeholders for future implementation);
		// these may be empty or contain only a "_reason" field.
		if len(bytes.TrimSpace(data)) == 0 {
			return nil
		}

		var t Template
		if err := json.Unmarshal(data, &t); err != nil {
			return fmt.Errorf("parsing template %s: %w", fpath, err)
		}

		if t.Name == "" {
			// Stub file with metadata (e.g. _reason) but no template definition; skip it
			return nil
		}

		if _, exists := r.templates[t.Name]; exists {
			return fmt.Errorf("duplicate template name %q in %s", t.Name, fpath)
		}

		// Resolve code_template_file references
		dir := path.Dir(fpath)
		for i := range t.Files {
			if t.Files[i].CodeTemplateFile == "" {
				continue
			}
			tmplPath := dir + "/" + t.Files[i].CodeTemplateFile
			content, err := definitionsFS.ReadFile(tmplPath)
			if err != nil {
				return fmt.Errorf("reading code template file %s for %s: %w", tmplPath, fpath, err)
			}
			t.Files[i].CodeTemplate = string(content)
		}

		// Preserve JSON-defined property order (Go maps lose insertion order).
		// Store on struct for backend use, and in the schema as []any for
		// protobuf Struct conversion ([]string is not protobuf-compatible).
		if t.JSONSchema != nil {
			t.PropertyOrder = extractPropertyOrder(data)
			if len(t.PropertyOrder) > 0 {
				orderAny := make([]any, len(t.PropertyOrder))
				for i, k := range t.PropertyOrder {
					orderAny[i] = k
				}
				t.JSONSchema["x-property-order"] = orderAny
			}
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
		// Combined templates (e.g. s3-duckdb) contain both connector and model files.
		// Check for a combined template first; fall back to standalone connector template.
		if t, ok := r.Get(driver + "-duckdb"); ok && hasFileNamed(t, "connector") {
			return t, true
		}
		return r.Get(driver)
	case "model":
		// Model templates use the pattern driver-{olap} (e.g. s3-duckdb, s3-clickhouse).
		// LookupByDriver defaults to DuckDB; for other OLAPs use Get() directly.
		return r.Get(driver + "-duckdb")
	}
	return nil, false
}

// hasFileNamed returns true if the template has a file entry with the given name.
func hasFileNamed(t *Template, name string) bool {
	for _, f := range t.Files {
		if f.Name == name {
			return true
		}
	}
	return false
}

// extractPropertyOrder parses raw JSON bytes to extract the key ordering of
// json_schema.properties. Go's map[string]any loses insertion order on unmarshal,
// but json.Decoder preserves it.
func extractPropertyOrder(raw []byte) []string {
	var outer struct {
		JSONSchema json.RawMessage `json:"json_schema"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil || outer.JSONSchema == nil {
		return nil
	}

	var schema struct {
		Properties json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(outer.JSONSchema, &schema); err != nil || schema.Properties == nil {
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(schema.Properties))
	tok, err := dec.Token() // opening {
	if err != nil {
		return nil
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil
	}

	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		key, ok := tok.(string)
		if !ok {
			break
		}
		keys = append(keys, key)
		// Skip the property value object
		var discard json.RawMessage
		if err := dec.Decode(&discard); err != nil {
			break
		}
	}
	return keys
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
