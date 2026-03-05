package templates

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// funcMap returns the template function map available to all template definitions.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"renderProps":    renderProps,
		"indent":         indent,
		"quote":          quote,
		"propVal":        propVal,
		"default":        defaultVal,
		"duckdbSQL":      duckdbSQL,
		"azureContainer": azureContainer,
		"azureBlobPath":  azureBlobPath,
	}
}

// renderProps renders a slice of ProcessedProp as YAML key-value lines.
// Each property is rendered on its own line with appropriate formatting:
// quoted values get double quotes, unquoted values are rendered as-is.
func renderProps(props []ProcessedProp) string {
	if len(props) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range props {
		if i > 0 {
			b.WriteByte('\n')
		}
		if p.Quoted {
			fmt.Fprintf(&b, "%s: %q", p.Key, p.Value)
		} else {
			fmt.Fprintf(&b, "%s: %s", p.Key, p.Value)
		}
	}
	return b.String()
}

// indent prepends each line of text with n spaces.
func indent(n int, text string) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = pad + line
		}
	}
	return strings.Join(lines, "\n")
}

// quote wraps a string in double quotes.
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

// propVal extracts a value from a []ProcessedProp by key.
// Returns "" if the key is not found or props is not the expected type.
func propVal(props any, key string) string {
	ps, ok := props.([]ProcessedProp)
	if !ok {
		return ""
	}
	for _, p := range ps {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// defaultVal returns val if non-empty, otherwise fallback.
// Registered as "default" in the template function map.
func defaultVal(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

// duckdbSQL maps a file path to a DuckDB read function call based on file extension.
// When defaultToJSON is true, unknown extensions default to read_json; otherwise select * from 'path'.
func duckdbSQL(path string, defaultToJSON bool) string {
	switch {
	case matchesExt(path, ".csv", ".tsv", ".txt"):
		return fmt.Sprintf("select * from read_csv('%s', auto_detect=true, ignore_errors=1, header=true)", path)
	case matchesExt(path, ".parquet"):
		return fmt.Sprintf("select * from read_parquet('%s')", path)
	case matchesExt(path, ".json", ".ndjson"):
		return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
	default:
		if defaultToJSON {
			return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
		}
		return fmt.Sprintf("select * from '%s'", path)
	}
}

// matchesExt checks if the file path has any of the target extensions.
// Handles compound extensions like .v1.parquet.gz by checking the basename
// for extension substrings, using suffix matching on path segments to avoid
// false positives (e.g. "parquet-archive/readme.txt" should NOT match ".parquet").
func matchesExt(path string, targets ...string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))
	for _, t := range targets {
		if ext == t {
			return true
		}
		// Check for compound extensions in the basename only (not the full path).
		// Look for the target extension followed by a dot or end-of-string.
		idx := strings.Index(base, t)
		if idx > 0 {
			after := idx + len(t)
			if after == len(base) || base[after] == '.' {
				return true
			}
		}
	}
	return false
}

// azureContainer extracts the container name from an Azure URI like "azure://container/blob/path".
func azureContainer(path string) string {
	path = strings.TrimPrefix(path, "azure://")
	idx := strings.IndexByte(path, '/')
	if idx < 0 {
		return path
	}
	return path[:idx]
}

// azureBlobPath extracts the blob path from an Azure URI like "azure://container/blob/path".
func azureBlobPath(path string) string {
	path = strings.TrimPrefix(path, "azure://")
	idx := strings.IndexByte(path, '/')
	if idx < 0 {
		return ""
	}
	return path[idx+1:]
}
