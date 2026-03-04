package templates

import "fmt"

// BuildClickHouseObjectStoreQuery generates a SELECT using ClickHouse's s3() or gcs() table function.
// Credential refs are Rill env var references (e.g. "{{ .env.AWS_ACCESS_KEY_ID }}").
func BuildClickHouseObjectStoreQuery(fn, path, keyRef, secretRef string) string {
	format := clickHouseFormat(path)

	sql := fmt.Sprintf("SELECT * FROM %s(\n    '%s'", fn, path)
	if keyRef != "" && secretRef != "" {
		sql += fmt.Sprintf(",\n    '%s',\n    '%s'", keyRef, secretRef)
	}
	if format != "" {
		sql += fmt.Sprintf(",\n    '%s'", format)
	}
	sql += "\n)"
	return sql
}

// BuildClickHouseAzureQuery generates a SELECT using azureBlobStorage().
func BuildClickHouseAzureQuery(endpoint, container, blobPath, accountRef, keyRef string) string {
	format := clickHouseFormat(blobPath)

	sql := fmt.Sprintf("SELECT * FROM azureBlobStorage(\n    '%s',\n    '%s',\n    '%s',\n    '%s',\n    '%s'",
		endpoint, container, blobPath, accountRef, keyRef)
	if format != "" {
		sql += fmt.Sprintf(",\n    '%s'", format)
	}
	sql += "\n)"
	return sql
}

// BuildClickHouseDatabaseQuery generates a SELECT using mysql() or postgresql().
func BuildClickHouseDatabaseQuery(fn, hostPort, database, table, user, password string) string {
	return fmt.Sprintf("SELECT * FROM %s(\n    '%s',\n    '%s',\n    '%s',\n    '%s',\n    '%s'\n)",
		fn, hostPort, database, table, user, password)
}

// BuildClickHouseURLQuery generates a SELECT using ClickHouse's url() function.
func BuildClickHouseURLQuery(url string) string {
	format := clickHouseFormat(url)
	if format == "" {
		format = "JSONEachRow"
	}
	return fmt.Sprintf("SELECT * FROM url(\n    '%s',\n    '%s'\n)", url, format)
}

// BuildClickHouseFileQuery generates a SELECT using ClickHouse's file() function.
func BuildClickHouseFileQuery(path string) string {
	format := clickHouseFormat(path)
	if format != "" {
		return fmt.Sprintf("SELECT * FROM file('%s', '%s')", path, format)
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
		return "Parquet"
	case matchesExt(path, ".csv"):
		return "CSVWithNames"
	case matchesExt(path, ".tsv", ".txt"):
		return "TabSeparated"
	case matchesExt(path, ".ndjson", ".json"):
		return "JSONEachRow"
	default:
		return ""
	}
}
