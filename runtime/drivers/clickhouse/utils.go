package clickhouse

func safeSQLName(name string) string {
	return DialectClickhouse.EscapeIdentifier(name)
}

func localTableName(name string) string {
	return name + "_local"
}
