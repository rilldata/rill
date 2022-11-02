const SingleQuoteRegex = /'/g;
const DoubleQuoteRegex = /"/g;

/**
 * Returns an escaped column surrounded by double quotes.
 * To be used when user entered column names are used in queries.
 * Any instances of " within the column name is replaced with "" to escape the "
 */
export function escapeColumn(columnName: string): string {
  return `"${columnName.replace(DoubleQuoteRegex, '""')}"`;
}

/**
 * Returns an escaped alias to be used for the column. Replaces all quotes with __
 */
export function escapeColumnAlias(columnName: string): string {
  return columnName
    .replace(SingleQuoteRegex, "__")
    .replace(DoubleQuoteRegex, "__");
}
