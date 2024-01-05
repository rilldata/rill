import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  MetricsViewFilterCond,
  V1Expression,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";

export function convertFilterToExpression(
  filter: V1MetricsViewFilter
): V1Expression {
  const exprs = new Array<V1Expression>();

  if (filter.include?.length) {
    const includeExprs = new Array<V1Expression>();
    for (const includeFilter of filter.include) {
      const expr = convertConditionToExpression(includeFilter, false);
      if (expr) {
        includeExprs.push(expr);
      }
    }
    if (includeExprs.length) {
      exprs.push(createOrExpression(includeExprs));
    }
  }

  if (filter.exclude?.length) {
    for (const excludeFilter of filter.exclude) {
      const expr = convertConditionToExpression(excludeFilter, true);
      if (expr) {
        exprs.push(expr);
      }
    }
  }

  return createAndExpression(exprs);
}

function convertConditionToExpression(
  cond: MetricsViewFilterCond,
  exclude: boolean
) {
  let inExpr: V1Expression | undefined;
  if (cond.in?.length) {
    inExpr = createInExpression(cond.name as string, cond.in, exclude);
  }

  const likeExpr = convertLikeToExpression(cond, exclude);

  if (inExpr && likeExpr) {
    return exclude
      ? createAndExpression([inExpr, likeExpr])
      : createOrExpression([inExpr, likeExpr]);
  } else if (inExpr) {
    return inExpr;
  } else if (likeExpr) {
    return likeExpr;
  }
  return undefined;
}

function convertLikeToExpression(
  cond: MetricsViewFilterCond,
  exclude: boolean
) {
  if (!cond.like?.length) return undefined;

  if (cond.like.length === 1) {
    return createLikeExpression(cond.name as string, cond.like[0], exclude);
  } else {
    const likeExprs = cond.like.map((v) =>
      createLikeExpression(cond.name as string, v, exclude)
    );
    return exclude
      ? createAndExpression(likeExprs)
      : createOrExpression(likeExprs);
  }
}
