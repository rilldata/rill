import {
  createAndExpression,
  isAndOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type MergedFilterParams = {
  expr: V1Expression;
  inList: string[];
  // True when the merged expression cannot be represented as chips.
  advanced: boolean;
};

/**
 * Merges the filter params of individual metrics views into a single expression.
 *
 * The UI has one filter editor, so the filters of every metrics view have to be folded into one
 * expression. Each param is parsed and its top level conditions are unioned, in the order the
 * metrics views are given. Conditions that repeat across metrics views are kept once, matched on
 * meaning rather than shape, so `country IN ('US','CA')` and `country IN ('CA','US')` merge.
 *
 * `advanced` reports that the merge cannot be shown as chips, because a condition is a nested
 * AND/OR, or because a dimension or measure ended up with more than one condition. The latter
 * happens when metrics views filter the same identifier differently: only one of those conditions
 * would get a chip, and editing it would leave the others in place.
 */
export function mergeFilterParams(
  paramByMetricsView: Record<string, string>,
): MergedFilterParams {
  const exprs: V1Expression[] = [];
  const seenExprs = new Set<string>();
  const seenIdentifiers = new Set<string>();
  const dimensionsWithInlistFilter: string[] = [];
  const seenInlistDimensions = new Set<string>();
  let advanced = false;

  for (const param of Object.values(paramByMetricsView)) {
    const parsed = parseFilterParam(param);

    for (const condition of topLevelConditions(parsed.expr)) {
      // A nested AND/OR has no chip of its own.
      if (isAndOrExpression(condition)) advanced = true;

      const key = canonicalKey(condition);
      if (seenExprs.has(key)) continue;
      seenExprs.add(key);
      exprs.push(condition);

      const identifier = getIdentifier(condition);
      if (identifier === undefined) continue;
      // Reached only for a condition that is not a duplicate,
      // so the identifier repeating means it is filtered two different ways.
      if (seenIdentifiers.has(identifier)) advanced = true;
      seenIdentifiers.add(identifier);
    }

    for (const dimension of parsed.dimensionsWithInlistFilter) {
      if (seenInlistDimensions.has(dimension)) continue;
      seenInlistDimensions.add(dimension);
      dimensionsWithInlistFilter.push(dimension);
    }
  }

  return {
    expr: createAndExpression(exprs),
    inList: dimensionsWithInlistFilter,
    advanced,
  };
}

/**
 * The conditions a param contributes to the merge.
 *
 * Only a top level AND is unwrapped. A top level OR joins its operands, so unwrapping it into the
 * merged AND would change what it matches; it goes in whole and is reported as advanced instead.
 */
function topLevelConditions(expr: V1Expression | undefined): V1Expression[] {
  if (!expr) return [];
  if (expr.cond?.op === V1Operation.OPERATION_AND) return expr.cond.exprs ?? [];
  return [expr];
}

/** The dimension or measure a condition filters, which is what its chip is keyed by. */
function getIdentifier(expr: V1Expression): string | undefined {
  // A measure filter is a subquery on a dimension, so its chip is keyed by the measure instead.
  const subquery = expr.cond?.exprs?.[1]?.subquery;
  if (subquery) return subquery.measures?.[0] ?? subquery.dimension;
  return expr.cond?.exprs?.[0]?.ident;
}

/**
 * A key that two conditions share when they filter the same way.
 *
 * Operands that the operation does not order, the values of an IN and the arms of an AND/OR,
 * are sorted, so conditions that differ only in how they were written still merge.
 */
function canonicalKey(expr: V1Expression | undefined): string {
  if (!expr) return "";

  if (expr.subquery) {
    const { dimension, measures, where, having } = expr.subquery;
    const parts = [
      dimension ?? "",
      measures?.join(",") ?? "",
      canonicalKey(where),
      canonicalKey(having),
    ];
    return `sub(${parts.join("|")})`;
  }

  if (expr.cond) {
    const op = expr.cond.op ?? V1Operation.OPERATION_UNSPECIFIED;
    const keys = (expr.cond.exprs ?? []).map(canonicalKey);
    switch (op) {
      case V1Operation.OPERATION_IN:
      case V1Operation.OPERATION_NIN:
        // The identifier leads, the values after it are a set.
        return `${op}(${[keys[0] ?? "", ...keys.slice(1).sort()].join(",")})`;
      case V1Operation.OPERATION_AND:
      case V1Operation.OPERATION_OR:
        return `${op}(${keys.sort().join(",")})`;
      default:
        return `${op}(${keys.join(",")})`;
    }
  }

  if (expr.ident !== undefined) return `id(${expr.ident})`;
  return `val(${JSON.stringify(expr.val ?? null)})`;
}

/**
 * The parser throws on a malformed param.
 * Converting the url to explore state reports it as an error, so drop it here instead of throwing.
 */
function parseFilterParam(param: string) {
  try {
    return convertFilterParamToExpression(param);
  } catch {
    return { expr: undefined, dimensionsWithInlistFilter: [] };
  }
}
