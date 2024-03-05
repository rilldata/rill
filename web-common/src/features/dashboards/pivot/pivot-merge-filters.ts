import {
  copyFilterExpression,
  createAndExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";

function valueIntersection(
  valueArray1: V1Expression[],
  valueArray2: V1Expression[],
) {
  return valueArray1.filter((obj1) =>
    valueArray2.some((obj2) => obj1.val === obj2.val),
  );
}

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

    /**
     * We take an intersection of the values in the IN expressions.
     * This is to make sure sorting, row expansion filter all work as
     * expected along with global filters. Otherwise, we would get data
     * for a larger subset than intended
     */
    const inExpr = inExprMap.get(ident);
    const inExprVals = inExpr?.cond?.exprs?.slice(1) ?? [];
    const exprVals = e.cond?.exprs?.slice(1) ?? [];
    const intersection = valueIntersection(inExprVals, exprVals);
    if (intersection.length === 0) {
      // no intersection, remove the identifier from the map
      inExprMap.delete(ident);
      return;
    }
    // replace the expression with the intersection
    e.cond!.exprs = [{ ident }, ...intersection]; // asserting that e.cond is not undefined
    // remove the identifier from the map
    inExprMap.delete(ident);
  });

  // add the remaining in expressions
  inExprMap.forEach((ie) => {
    if (!filter2.cond?.exprs) {
      filter2.cond!.exprs = [copyFilterExpression(ie)];
    }
    filter2.cond?.exprs?.push(copyFilterExpression(ie));
  });
  // add all like expressions
  likeExprMap.forEach((ie) => {
    if (!filter2.cond?.exprs) {
      filter2.cond!.exprs = [copyFilterExpression(ie)];
    }
    filter2.cond?.exprs?.push(copyFilterExpression(ie));
  });

  return filter2;
}
