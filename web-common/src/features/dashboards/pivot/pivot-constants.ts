/**
 * Allowed row limit values for the pivot table.
 * These values are used both in the UI dropdown and for URL validation.
 * undefined represents unlimited rows (displayed as "All" in the UI)
 */
export const PIVOT_ROW_LIMIT_OPTIONS = [5, 10, 25, 50, 100] as const;

/**
 * Calculates the effective row limit to apply for a query based on the configured
 * row limit, current offset, and page size.
 *
 * @param rowLimit - The maximum number of rows to fetch (undefined = unlimited)
 * @param rowOffset - The current row offset (for pagination)
 * @param pageSize - The number of rows per page
 * @returns The limit to apply as a string for the query
 */
export function calculateEffectiveRowLimit(
  rowLimit: number | undefined,
  rowOffset: number,
  pageSize: number,
): string {
  if (rowLimit === undefined) {
    return pageSize.toString();
  }
  const remainingRows = rowLimit - rowOffset;
  if (remainingRows <= 0) {
    return "0";
  }
  return Math.min(remainingRows, pageSize).toString();
}
