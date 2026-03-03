package templates

import (
	"fmt"
	"strings"
)

// ClickHouse format names for table functions.
const (
	chFormatCSV     = "'CSVWithNames'"
	chFormatTSV     = "'TabSeparated'"
	chFormatParquet = "'Parquet'"
	chFormatJSON    = "'JSONEachRow'"
)

// BuildClickHouseObjectStoreQuery generates a SELECT using ClickHouse's s3(), gcs(), or
// azureBlobStorage() table function. Credential refs should be Rill env var references
// (e.g. "{{ .env.AWS_ACCESS_KEY_ID }}") or literal values.
func BuildClickHouseObjectStoreQuery(fn, path, keyRef, secretRef string) string {
	format := clickHouseFormat(path)

	var sb strings.Builder
	fmt.Fprintf(&sb, "SELECT * FROM %s(\n", fn)
	fmt.Fprintf(&sb, "    '%s'", path)

	if keyRef != "" && secretRef != "" {
		fmt.Fprintf(&sb, ",\n    '%s'", keyRef)
		fmt.Fprintf(&sb, ",\n    '%s'", secretRef)
	}

	if format != "" {
		fmt.Fprintf(&sb, ",\n    %s", format)
	}

	sb.WriteString("\n)")
	return sb.String()
}

// BuildClickHouseAzureQuery generates a SELECT using azureBlobStorage().
// Accepts endpoint, container, blobPath, account, and key/SAS reference.
func BuildClickHouseAzureQuery(endpoint, container, blobPath, accountRef, keyRef string) string {
	format := clickHouseFormat(blobPath)

	var sb strings.Builder
	sb.WriteString("SELECT * FROM azureBlobStorage(\n")
	fmt.Fprintf(&sb, "    '%s',\n", endpoint)
	fmt.Fprintf(&sb, "    '%s',\n", container)
	fmt.Fprintf(&sb, "    '%s',\n", blobPath)
	fmt.Fprintf(&sb, "    '%s',\n", accountRef)
	fmt.Fprintf(&sb, "    '%s'", keyRef)

	if format != "" {
		fmt.Fprintf(&sb, ",\n    %s", format)
	}

	sb.WriteString("\n)")
	return sb.String()
}

// BuildClickHouseDatabaseQuery generates a SELECT using mysql() or postgresql().
func BuildClickHouseDatabaseQuery(fn, hostPort, database, table, user, password string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "SELECT * FROM %s(\n", fn)
	fmt.Fprintf(&sb, "    '%s',\n", hostPort)
	fmt.Fprintf(&sb, "    '%s',\n", database)
	fmt.Fprintf(&sb, "    '%s',\n", table)
	fmt.Fprintf(&sb, "    '%s',\n", user)
	fmt.Fprintf(&sb, "    '%s'\n", password)
	sb.WriteString(")")
	return sb.String()
}

// BuildClickHouseURLQuery generates a SELECT using ClickHouse's url() function.
func BuildClickHouseURLQuery(url string) string {
	format := clickHouseFormat(url)
	if format == "" {
		format = chFormatJSON // default to JSONEachRow for URLs
	}
	return fmt.Sprintf("SELECT * FROM url(\n    '%s',\n    %s\n)", url, format)
}

// BuildClickHouseFileQuery generates a SELECT using ClickHouse's file() function.
func BuildClickHouseFileQuery(path string) string {
	format := clickHouseFormat(path)
	if format != "" {
		return fmt.Sprintf("SELECT * FROM file('%s', %s)", path, format)
	}
	return fmt.Sprintf("SELECT * FROM file('%s')", path)
}

// BuildClickHouseSQLiteQuery generates a SELECT using ClickHouse's sqlite() function.
func BuildClickHouseSQLiteQuery(dbPath, table string) string {
	return fmt.Sprintf("SELECT * FROM sqlite('%s', '%s')", dbPath, table)
}

// clickHouseFormat infers the ClickHouse format name from a file path's extension.
func clickHouseFormat(path string) string {
	switch {
	case matchesExt(path, ".parquet"):
		return chFormatParquet
	case matchesExt(path, ".csv"):
		return chFormatCSV
	case matchesExt(path, ".tsv", ".txt"):
		return chFormatTSV
	case matchesExt(path, ".ndjson", ".json"):
		return chFormatJSON
	default:
		return ""
	}
}
