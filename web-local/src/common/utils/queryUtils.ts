export function sanitizeColumn(columnName: string) {
  // // special-case for how duckdb works
  // if (columnName === 'count(*)') { return '"count_star()"'};
  return `"${columnName}"`;
}
