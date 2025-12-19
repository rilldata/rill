export const SHOW_MORE_BUTTON = "__rill_type_SHOW_MORE_BUTTON";
export const LOADING_CELL = "__rill_type_LOADING_CELL";

export const MAX_ROW_EXPANSION_LIMIT = 100;

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

/**
 * Gets the next limit in the progression: 5 → 10 → 25 → 50 → 100
 * If current limit is not in the standard progression, returns the next higher value.
 *
 * @param currentLimit - The current row limit
 * @returns The next limit in the progression, or undefined if at/beyond 100
 */
export function getNextRowLimit(currentLimit: number): number | undefined {
  const limits = [...PIVOT_ROW_LIMIT_OPTIONS]; // [5, 10, 25, 50, 100]
  return limits.find((limit) => limit > currentLimit) ?? undefined;
}

/**
 * Gets the display label for the next limit in the progression.
 *
 * @param currentLimit - The current row limit
 * @returns The label to display (e.g., "10", "25", "100")
 */
export function getNextLimitLabel(currentLimit: number): string {
  const nextLimit = getNextRowLimit(currentLimit);
  return nextLimit ? nextLimit.toString() : "100";
}
