package clickhouse

func safeSQLName(name string) string {
	return newDialect().EscapeIdentifier(name)
}
