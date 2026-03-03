package templates

// Template is a declarative definition for generating project files (connectors, models, etc.)
// from structured form data. Each template knows how to produce one or more output files.
type Template struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Driver      string   `yaml:"driver"` // primary driver (e.g. "s3"); empty for driverless templates like iceberg
	OLAP        string   `yaml:"olap"`   // target OLAP engine (e.g. "duckdb"); empty for OLAP connector templates
	Tags        []string `yaml:"tags"`
	Files       []File   `yaml:"files"`
}

// File describes a single output file within a template.
// PathTemplate and CodeTemplate use Go text/template syntax with [[ ]] delimiters
// to avoid collision with Rill's {{ .env.VAR }} runtime syntax.
type File struct {
	Name         string `yaml:"name"`          // output name: "connector" or "model"
	PathTemplate string `yaml:"path_template"` // Go template for the file path
	CodeTemplate string `yaml:"code_template"` // Go template for the file content
}

// ProcessedProp is a property that has been pre-processed for template rendering.
// Secret values are replaced with {{ .env.VAR }} references; empty values are filtered.
type ProcessedProp struct {
	Key    string
	Value  string
	Quoted bool // true for strings and secrets; false for numbers and booleans
}
