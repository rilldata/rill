import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";

export interface ExtractedFilter {
  dimensionName: string;
  values: string[];
}

/**
 * Extracts dimension filters from a V1Expression structure.
 * The expression is expected to be an AND operation containing IN operations.
 *
 * @param filters - The V1Expression containing dimension filters
 * @returns Array of extracted dimension filters with their values
 */
export function extractDimensionFiltersFromExpression(
  filters: V1Expression | undefined,
): ExtractedFilter[] {
  if (!filters?.cond?.exprs) return [];

  const result: ExtractedFilter[] = [];

  // Walk the AND expression tree
  for (const expr of filters.cond.exprs) {
    if (expr.cond?.op === V1Operation.OPERATION_IN) {
      const ident = expr.cond.exprs?.[0]?.ident;
      if (!ident) continue;

      // Extract values (skip first expr which is the identifier)
      const values = expr?.cond?.exprs
        ?.slice(1)
        .map((e) => e.val)
        .filter((val): val is string => val !== undefined && val !== null);

      if (values && values?.length > 0) {
        result.push({ dimensionName: ident, values });
      }
    }
  }

  return result;
}
