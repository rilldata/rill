package templates

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BuildDuckDBQuery maps a file path + extension to a DuckDB read function call.
func BuildDuckDBQuery(path string, defaultToJSON bool) string {
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
