import {
  copyFilterExpression,
  createAndExpression,
  forEachIdentifier,
  wrapNonJoinerExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";

export function mergeFilters(
  filter1: V1Expression | undefined,
  filter2: V1Expression | undefined,
): V1Expression | undefined {
  if (!filter1 && !filter2) return undefined;
  if (!filter1) return filter2;
  if (!filter2) return filter1;
  filter1 = wrapNonJoinerExpression(filter1);
  filter2 = wrapNonJoinerExpression(filter2);

  // IN and NIN are merged separately. So maintain separate references
  const inExprMap = new Map<string, V1Expression>();
  const notInExprMap = new Map<string, V1Expression>();
  const likeExprMap = new Map<string, V1Expression>();

  // build a map of identifier to IN and LIKE expressions separately
  forEachIdentifier(filter1, (e, ident) => {
    if (
      e.cond?.op === V1Operation.OPERATION_LIKE ||
      e.cond?.op === V1Operation.OPERATION_NLIKE
    ) {
      if (likeExprMap.has(ident)) return;
      likeExprMap.set(ident, e);
    } else if (e.cond?.op === V1Operation.OPERATION_IN) {
      if (inExprMap.has(ident) || !!e.cond?.exprs?.[1]?.subquery) return;
      inExprMap.set(ident, e);
    } else if (e.cond?.op === V1Operation.OPERATION_NIN) {
      if (notInExprMap.has(ident) || !!e.cond?.exprs?.[1]?.subquery) return;
      notInExprMap.set(ident, e);
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
    if (e.cond?.exprs?.[1]?.subquery) return;

    if (inExprMap.has(ident) && e.cond?.op === V1Operation.OPERATION_IN) {
      /**
       * We take an intersection of the values in the IN expressions.
       * This is to make sure sorting, row expansion filter all work as
       * expected along with global filters. Otherwise, we would get data
       * for a larger subset than intended
       */
      const inExpr = inExprMap.get(ident)!;
      // replace the expression with the intersection
      e.cond.exprs = [{ ident }, ...valueIntersection(inExpr, e)];
      // remove the identifier from the map
      inExprMap.delete(ident);
    } else if (
      notInExprMap.has(ident) &&
      e.cond?.op === V1Operation.OPERATION_NIN
    ) {
      const notInExpr = notInExprMap.get(ident)!;
      // replace the expression with the union
      e.cond.exprs = [{ ident }, ...valueUnion(notInExpr, e)];
      // remove the identifier from the map
      notInExprMap.delete(ident);
    }
  });

  // add the remaining in expressions
  inExprMap.forEach((ie) => {
    if (!filter2.cond?.exprs) {
      filter2.cond!.exprs = [copyFilterExpression(ie)];
    }
    filter2.cond?.exprs?.push(copyFilterExpression(ie));
  });
  // add the remaining in expressions
  notInExprMap.forEach((ie) => {
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

function valueIntersection(inExpr1: V1Expression, inExpr2: V1Expression) {
  const inExpr1Vals = inExpr1?.cond?.exprs?.slice(1) ?? [];
  const inExpr2Vals = inExpr2.cond?.exprs?.slice(1) ?? [];

  const intersection = inExpr1Vals.filter((obj1) =>
    inExpr2Vals.some((obj2) => obj1.val === obj2.val),
  );
  // backwards compatibility. if there are no intersections, retain from expr1
  if (intersection.length === 0) return inExpr1Vals;
  return intersection;
}

function valueUnion(notInExpr1: V1Expression, notInExpr2: V1Expression) {
  const notInExpr1Vals = notInExpr1?.cond?.exprs?.slice(1) ?? [];
  const notInExpr2Vals = notInExpr2.cond?.exprs?.slice(1) ?? [];

  const seen = new Set(notInExpr1Vals.map((o1) => o1.val));
  const unionValues = [...notInExpr1Vals];

  notInExpr2Vals.forEach((o2) => {
    if (seen.has(o2.val)) return;
    unionValues.push(o2);
    seen.add(o2.val);
  });

  return unionValues;
}
