/**
 * Merge two filters together.
 * This might change later when move to the newer
 * filter format.
 */

import {
  copyFilterExpression,
  createAndExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export function mergeFilters(
  filter1: V1Expression,
  filter2: V1Expression,
): V1Expression {
  const inExprMap = new Map<string, V1Expression>();
  const likeExprMap = new Map<string, V1Expression>();

  // build a map of identifier to IN and LIKE expressions separately
  forEachIdentifier(filter1, (e, ident) => {
    if (
      e.cond?.op === V1Operation.OPERATION_LIKE ||
      e.cond?.op === V1Operation.OPERATION_NLIKE
    ) {
      if (likeExprMap.has(ident)) return;
      likeExprMap.set(ident, e);
    } else {
      if (inExprMap.has(ident)) return;
      inExprMap.set(ident, e);
    }
  });

  // create a copy
  filter2 = copyFilterExpression(filter2) ?? createAndExpression([]);
  forEachIdentifier(filter2, (e, ident) => {
    // ignore like expressions since those need individual expressions and cannot be merged
    if (
      e.cond?.op === V1Operation.OPERATION_LIKE ||
      e.cond?.op === V1Operation.OPERATION_NLIKE
    )
      return;
    if (!inExprMap.has(ident)) return;
    e.cond?.exprs?.push(...(inExprMap.get(ident)?.cond?.exprs?.slice(1) ?? []));
    inExprMap.delete(ident);
  });

  // add the remaining in expressions
  inExprMap.forEach((ie) => {
    filter2.cond?.exprs?.push(copyFilterExpression(ie));
  });
  // add all like expressions
  likeExprMap.forEach((ie) => {
    filter2.cond?.exprs?.push(copyFilterExpression(ie));
  });

  return filter2;
}
