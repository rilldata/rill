const SingleQuoteRegex = /'/g;
const DoubleQuoteRegex = /"/g;

export function escapeColumn(columnName: string): string {
  return `"${columnName.replace(DoubleQuoteRegex, '""')}"`;
}

export function escapeColumnAlias(columnName: string): string {
  return columnName
    .replace(SingleQuoteRegex, "__")
    .replace(DoubleQuoteRegex, "__");
}
