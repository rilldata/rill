package templates

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// sharedFuncMap is the template function map available to all template definitions.
// Allocated once since all entries are stateless functions.
var sharedFuncMap = template.FuncMap{
	"renderProps":         renderProps,
	"indent":              indent,
	"quote":               quote,
	"propVal":             propVal,
	"default":             defaultVal,
	"duckdbSQL":           duckdbSQL,
	"s3ToHTTPS":           s3ToHTTPS,
	"gcsToHTTPS":          gcsToHTTPS,
	"azureContainer":      azureContainer,
	"azureBlobPath":       azureBlobPath,
	"azureEndpoint":       azureEndpoint,
	"clickhouseFormat":    clickhouseFormat,
	"clickhouseURLSuffix": clickhouseURLSuffix,
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
// NOTE: Use positional syntax only: [[ default (expr) "fallback" ]].
// Pipeline syntax ([[ expr | default "fallback" ]]) would swap arguments
// because text/template pipes into the last parameter.
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

// s3ToHTTPS converts an s3:// URI to an HTTPS URL for ClickHouse's s3() function.
// "s3://bucket/key" becomes "https://bucket.s3.amazonaws.com/key".
// If the path is already HTTPS, it is returned as-is.
func s3ToHTTPS(path string) string {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return path
	}
	trimmed := strings.TrimPrefix(path, "s3://")
	idx := strings.IndexByte(trimmed, '/')
	if idx < 0 {
		return fmt.Sprintf("https://%s.s3.amazonaws.com", trimmed)
	}
	bucket := trimmed[:idx]
	key := trimmed[idx:]
	return fmt.Sprintf("https://%s.s3.amazonaws.com%s", bucket, key)
}

// gcsToHTTPS converts a gs:// URI to an HTTPS URL for ClickHouse's gcs() function.
// "gs://bucket/key" becomes "https://storage.googleapis.com/bucket/key".
// If the path is already HTTPS, it is returned as-is.
func gcsToHTTPS(path string) string {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return path
	}
	trimmed := strings.TrimPrefix(path, "gs://")
	return fmt.Sprintf("https://storage.googleapis.com/%s", trimmed)
}

// clickhouseURLSuffix generates additional arguments for ClickHouse's url() table function.
// When headers are present, returns ", Format, headers('K1'='V1', 'K2'='V2')" where
// Format is auto-detected from the URL extension. Returns empty string when no headers
// are present (ClickHouse auto-detects everything from the URL).
func clickhouseURLSuffix(path string, props any) string {
	hdrs := extractClickhouseHeaders(props)
	if hdrs == "" {
		return ""
	}
	format := clickhouseFormat(path)
	return fmt.Sprintf(",\n  %s,\n  %s\n  ", format, hdrs)
}

// clickhouseFormat maps a URL path to a ClickHouse input format name.
func clickhouseFormat(path string) string {
	switch {
	case matchesExt(path, ".csv", ".txt"):
		return "CSVWithNames"
	case matchesExt(path, ".tsv"):
		return "TabSeparatedWithNames"
	case matchesExt(path, ".json", ".ndjson", ".jsonl"):
		return "JSONEachRow"
	case matchesExt(path, ".parquet"):
		return "Parquet"
	default:
		return "JSONEachRow"
	}
}

// extractClickhouseHeaders extracts header ProcessedProps and formats them
// for ClickHouse's headers() syntax: headers('Key1'='value1', 'Key2'='value2').
// Returns empty string if no headers are present.
func extractClickhouseHeaders(props any) string {
	ps, ok := props.([]ProcessedProp)
	if !ok {
		return ""
	}
	var headerProp *ProcessedProp
	for i := range ps {
		if ps[i].Key == "headers" {
			headerProp = &ps[i]
			break
		}
	}
	if headerProp == nil {
		return ""
	}

	// Parse the YAML-style header lines (e.g. "Authorization: \"Bearer {{ .env.X }}\"")
	var pairs []string
	for _, line := range strings.Split(headerProp.Value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.IndexByte(line, ':')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, "\"")
		pairs = append(pairs, fmt.Sprintf("'%s'='%s'", key, val))
	}
	if len(pairs) == 0 {
		return ""
	}
	return fmt.Sprintf("headers(%s)", strings.Join(pairs, ", "))
}

// azureContainer extracts the container name from an Azure URI.
// Supports both "azure://container/blob/path" and "https://account.blob.core.windows.net/container/blob/path".
func azureContainer(path string) string {
	path = stripAzurePrefix(path)
	idx := strings.IndexByte(path, '/')
	if idx < 0 {
		return path
	}
	return path[:idx]
}

// azureBlobPath extracts the blob path from an Azure URI.
// Supports both "azure://container/blob/path" and "https://account.blob.core.windows.net/container/blob/path".
func azureBlobPath(path string) string {
	path = stripAzurePrefix(path)
	idx := strings.IndexByte(path, '/')
	if idx < 0 {
		return ""
	}
	return path[idx+1:]
}

// azureEndpoint returns the blob service endpoint for a given Azure URI.
// For "https://account.blob.core.windows.net/..." it returns "https://account.blob.core.windows.net".
// For "azure://..." it builds the endpoint from the account property.
func azureEndpoint(path, account string) string {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		u := strings.TrimPrefix(path, "https://")
		u = strings.TrimPrefix(u, "http://")
		idx := strings.IndexByte(u, '/')
		if idx > 0 {
			return "https://" + u[:idx]
		}
		return "https://" + u
	}
	return fmt.Sprintf("https://%s.blob.core.windows.net", account)
}

// stripAzurePrefix strips the scheme and host from an Azure URI, returning "container/blob/path".
// Handles "azure://container/path" and "https://account.blob.core.windows.net/container/path".
func stripAzurePrefix(path string) string {
	if strings.HasPrefix(path, "azure://") {
		return strings.TrimPrefix(path, "azure://")
	}
	// HTTPS: strip scheme + host, leaving "/container/path"
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		noScheme := strings.TrimPrefix(path, "https://")
		noScheme = strings.TrimPrefix(noScheme, "http://")
		idx := strings.IndexByte(noScheme, '/')
		if idx >= 0 {
			return noScheme[idx+1:]
		}
		return ""
	}
	return path
}
