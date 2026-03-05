package templates

// Template is a declarative definition for generating project files (connectors, models, etc.)
// from structured form data. Each template knows how to produce one or more output files.
//
// Templates with JSONSchema are self-contained: the schema drives form rendering (frontend)
// and property metadata like secret detection (backend). Templates without JSONSchema fall
// back to drivers.Spec for property metadata.
type Template struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description,omitempty"` // short description for UI display
	DocsURL     string         `json:"docs_url,omitempty"`    // link to documentation
	Driver      string         `json:"driver"`                // primary driver (e.g. "s3"); empty for driverless templates like iceberg
	OLAP        string         `json:"olap"`                  // target OLAP engine (e.g. "duckdb"); empty for OLAP connector templates
	Tags        []string       `json:"tags"`
	JSONSchema    map[string]any `json:"json_schema,omitempty"` // JSON Schema for form generation and property metadata
	PropertyOrder []string       `json:"-"`                     // JSON-defined property key order; computed at load time
	Files         []File         `json:"files"`
}

// File describes a single output file within a template.
// PathTemplate and CodeTemplate use Go text/template syntax with [[ ]] delimiters
// to avoid collision with Rill's {{ .env.VAR }} runtime syntax.
//
// CodeTemplate can be specified inline (code_template) or loaded from a separate file
// (code_template_file) for readability. If both are set, code_template_file wins.
type File struct {
	Name             string `json:"name"`                         // output name: "connector" or "model"
	PathTemplate     string `json:"path_template"`                // Go template for the file path
	CodeTemplate     string `json:"code_template,omitempty"`      // Go template for the file content (inline)
	CodeTemplateFile string `json:"code_template_file,omitempty"` // path to .tmpl file (relative to the JSON definition)
}

// ProcessedProp is a property that has been pre-processed for template rendering.
// Secret values are replaced with {{ .env.VAR }} references; empty values are filtered.
type ProcessedProp struct {
	Key    string
	Value  string
	Quoted bool // true for strings and secrets; false for numbers and booleans
}
