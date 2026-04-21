package clickhouse

func safeSQLName(name string) string {
	return DialectClickhouse.EscapeIdentifier(name)
}
