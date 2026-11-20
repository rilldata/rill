import {
  createAndExpression,
  createInExpression,
  getValuesInExpression,
  isAndOrExpression,
  isSubqueryExpression,
  maybeConvertEqualityToInExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  convertExpressionToFilterParam,
  convertFilterParamToExpression,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
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
 * The UI has one filter editor, so the filters of every metrics view have to be folded into one expression.
 * Each param is parsed and its top level conditions are unioned.
 * Conditions that repeat across metrics views are kept once,
 * matched on meaning rather than shape, so `country IN ('US','CA')` and `country IN ('CA','US')` merge.
 *
 * `advanced` reports that the merge cannot be shown as chips because,
 * 1. A condition is a nested AND/OR
 * 2. A condition uses an operation the chips have no mode for, `publisher GT 'x'` for example
 * 3. A dimension or measure ended up with more than one condition.
 * The last happens when metrics views filter the same identifier differently.
 */
export function mergeFilterParams(
  urlParams: URLSearchParams,
): MergedFilterParams {
  const exprs: V1Expression[] = [];
  const seenExprs = new Set<string>();
  const seenIdentifiers = new Set<string>();
  const dimensionsWithInlistFilter: string[] = [];
  const seenInlistDimensions = new Set<string>();
  let advanced = false;

  for (const param of urlParams.values()) {
    const parsed = parseFilterParam(param);

    for (const condition of topLevelConditions(parsed.expr)) {
      // A nested AND/OR has no chip of its own, and neither does an operation outside the chip modes.
      if (isAndOrExpression(condition) || !hasChipOperation(condition)) {
        advanced = true;
      }

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
 * Narrows the top level conditions on single-select dimensions to the one value each can hold,
 * the last value of an IN or NOT IN, included. A condition without such a value, a contains or a
 * comparison for example, is dropped. Nested AND/ORs are left alone, since they have no chips.
 *
 * `DimensionFilterManager` narrows its own state the same way. This is for filters read before a
 * manager holds them, so that the stored params and the explore state match the chips.
 * Returns undefined when nothing needed narrowing.
 */
export function narrowSingleSelectFilters(
  expr: V1Expression,
  isSingleSelect: (dimension: string) => boolean,
): V1Expression | undefined {
  let narrowed = false;
  const conditions: V1Expression[] = [];
  for (const condition of topLevelConditions(expr)) {
    const narrowedCondition = narrowSingleSelectCondition(
      condition,
      isSingleSelect,
    );
    if (narrowedCondition !== condition) narrowed = true;
    if (narrowedCondition) conditions.push(narrowedCondition);
  }
  return narrowed ? createAndExpression(conditions) : undefined;
}

/**
 * `narrowSingleSelectFilters` for a filter param.
 * Returns the narrowed param, empty when nothing is left, or undefined when nothing needed narrowing.
 */
export function narrowSingleSelectFilterParam(
  param: string,
  isSingleSelect: (dimension: string) => boolean,
): string | undefined {
  const { expr, dimensionsWithInlistFilter } = parseFilterParam(param);
  const narrowed = expr && narrowSingleSelectFilters(expr, isSingleSelect);
  if (!narrowed) return undefined;
  // A single-select filter is never in "In List" mode.
  return convertExpressionToFilterParam(
    narrowed,
    dimensionsWithInlistFilter.filter((d) => !isSingleSelect(d)),
  );
}

/** Returns the condition itself when it needs no narrowing, and undefined when it is dropped. */
function narrowSingleSelectCondition(
  condition: V1Expression,
  isSingleSelect: (dimension: string) => boolean,
): V1Expression | undefined {
  const ident = condition.cond?.exprs?.[0]?.ident;
  // A measure filter is a subquery keyed by a dimension, not a filter on its values.
  if (!ident || !isSingleSelect(ident) || isSubqueryExpression(condition)) {
    return condition;
  }

  // `=` and `!=` are a single value selection and its exclude form.
  const inExpr = maybeConvertEqualityToInExpressions(condition);
  const op = inExpr.cond?.op;
  if (op !== V1Operation.OPERATION_IN && op !== V1Operation.OPERATION_NIN) {
    return undefined;
  }
  const values = getValuesInExpression(inExpr);
  if (op === V1Operation.OPERATION_IN && values.length === 1) return condition;
  return values.length
    ? createInExpression(ident, values.slice(-1))
    : undefined;
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

/**
 * Whether a chip can show the condition's operation.
 * A measure filter is `dimension IN <subquery>`, so it passes along with the dimension modes.
 * `=` and `!=` pass as a single value selection, see `DimensionFilterManager.reconcile`.
 */
function hasChipOperation(expr: V1Expression) {
  switch (expr.cond?.op) {
    case V1Operation.OPERATION_IN:
    case V1Operation.OPERATION_NIN:
    case V1Operation.OPERATION_LIKE:
    case V1Operation.OPERATION_NLIKE:
      return true;
    case V1Operation.OPERATION_EQ:
    case V1Operation.OPERATION_NEQ:
      return (expr.cond.exprs ?? []).slice(1).every((e) => "val" in e);
    default:
      return false;
  }
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
