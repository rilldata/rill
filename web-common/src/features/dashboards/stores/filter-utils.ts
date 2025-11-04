import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  V1Operation,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1Condition,
  type V1Expression,
  type V1Subquery,
} from "@rilldata/web-common/runtime-client";

export function createLikeExpression(
  ident: string,
  like: string,
  negate = false,
): V1Expression {
  return {
    cond: {
      op: negate ? V1Operation.OPERATION_NLIKE : V1Operation.OPERATION_LIKE,
      exprs: [{ ident }, { val: like }],
    },
  };
}

export function createInExpression(
  ident: string,
  vals: any[],
  negate = false,
): V1Expression {
  return {
    cond: {
      op: negate ? V1Operation.OPERATION_NIN : V1Operation.OPERATION_IN,
      exprs: [{ ident }, ...vals.map((val) => ({ val }))],
    },
  };
}

export function createAndExpression(exprs: V1Expression[]): V1Expression {
  return {
    cond: {
      op: V1Operation.OPERATION_AND,
      exprs,
    },
  };
}

export function createOrExpression(exprs: V1Expression[]): V1Expression {
  return {
    cond: {
      op: V1Operation.OPERATION_OR,
      exprs,
    },
  };
}

export function createBinaryExpression(
  ident: string,
  op: V1Operation,
  val: any,
): V1Expression {
  return {
    cond: {
      op,
      exprs: [{ ident }, { val }],
    },
  };
}

export function createBetweenExpression(
  ident: string,
  val1: number,
  val2: number,
  negate: boolean,
): V1Expression {
  const exprs: V1Expression[] = [
    {
      cond: {
        op: negate ? V1Operation.OPERATION_LTE : V1Operation.OPERATION_GT,
        exprs: [{ ident }, { val: val1 }],
      },
    },
    {
      cond: {
        op: negate ? V1Operation.OPERATION_GTE : V1Operation.OPERATION_LT,
        exprs: [{ ident }, { val: val2 }],
      },
    },
  ];
  if (negate) {
    return createOrExpression(exprs);
  } else {
    return createAndExpression(exprs);
  }
}

export function isBetweenExpression(expr: V1Expression): boolean {
  if (!expr.cond || expr.cond.exprs?.length !== 2) return false;

  const isBetween =
    expr.cond.op === V1Operation.OPERATION_AND &&
    expr.cond.exprs[0].cond?.op === V1Operation.OPERATION_GT &&
    expr.cond.exprs[1].cond?.op === V1Operation.OPERATION_LT;

  const isNotBetween =
    expr.cond.op === V1Operation.OPERATION_OR &&
    expr.cond.exprs[0].cond?.op === V1Operation.OPERATION_LTE &&
    expr.cond.exprs[1].cond?.op === V1Operation.OPERATION_GTE;

  return isBetween || isNotBetween;
}

export function createSubQueryExpression(
  dimension: string,
  measures: string[],
  having: V1Expression | undefined,
): V1Expression {
  return {
    cond: {
      op: V1Operation.OPERATION_IN,
      exprs: [
        { ident: dimension },
        {
          subquery: {
            dimension,
            measures,
            having,
          },
        },
      ],
    },
  };
}

const conditionOperationComplement: Partial<Record<V1Operation, V1Operation>> =
  {
    [V1Operation.OPERATION_EQ]: V1Operation.OPERATION_NEQ,
    [V1Operation.OPERATION_LT]: V1Operation.OPERATION_GTE,
    [V1Operation.OPERATION_LTE]: V1Operation.OPERATION_GT,
    [V1Operation.OPERATION_IN]: V1Operation.OPERATION_NIN,
    [V1Operation.OPERATION_LIKE]: V1Operation.OPERATION_NLIKE,
    [V1Operation.OPERATION_AND]: V1Operation.OPERATION_OR,
  };
// add inverse of existing values above
for (const c in conditionOperationComplement) {
  conditionOperationComplement[conditionOperationComplement[c]] = c;
}

export function negateExpression(expr: V1Expression): V1Expression {
  if ("ident" in expr || "val" in expr || !expr.cond) return expr;
  return {
    cond: {
      op:
        conditionOperationComplement[expr.cond.op as V1Operation] ??
        V1Operation.OPERATION_EQ,
      exprs: expr.cond.exprs,
    },
  };
}

export function forEachExpression(
  expr: V1Expression,
  cb: (e: V1Expression, depth: number) => void,
  depth = 0,
) {
  cb(expr, depth);
  if (!expr.cond?.exprs) return;

  for (const subExpr of expr.cond.exprs) {
    forEachExpression(subExpr, cb, depth + 1);
  }
}

export function forEachIdentifier(
  expr: V1Expression,
  cb: (e: V1Expression, ident: string) => void,
) {
  forEachExpression(expr, (e) => {
    const ident = e.cond?.exprs?.[0]?.ident;
    if (ident === undefined) {
      return;
    }
    cb(e, ident);
  });
}

export function getAllIdentifiers(expr: V1Expression | undefined): string[] {
  if (!expr) return [];
  const idents = new Set<string>();
  forEachExpression(expr, (e) => {
    if (e.ident) {
      idents.add(e.ident);
    }
  });
  return [...idents];
}

/**
 * Creates a copy of the expression with sub expressions filtered based on {@link checker}
 */
export function filterExpressions(
  expr: V1Expression,
  checker: (e: V1Expression) => boolean,
): V1Expression | undefined {
  if (expr.subquery) {
    const newSubquery = filterSubQuery(expr.subquery, checker);
    if (!newSubquery.having && !newSubquery.where) return undefined;
    return {
      subquery: newSubquery,
    };
  }

  if (!expr.cond?.exprs) {
    return {
      ...expr,
    };
  }

  const newExpr: V1Expression = {
    cond: {
      op: expr.cond.op,
      exprs: expr.cond.exprs
        .map((e) => filterExpressions(e, checker))
        .filter((e) => e !== undefined && checker(e)) as V1Expression[],
    },
  };

  switch (expr.cond.op) {
    // and/or will have only sub expressions
    case V1Operation.OPERATION_AND:
    case V1Operation.OPERATION_OR:
      if (newExpr.cond?.exprs?.length === 0) return undefined;
      break;

    default:
      // other types should have at least 2 expressions
      if (newExpr.cond?.exprs?.length && newExpr.cond.exprs.length <= 1)
        return undefined;
      break;
  }

  return newExpr;
}
function filterSubQuery(
  subQuery: V1Subquery,
  checker: (e: V1Expression) => boolean,
) {
  if (subQuery.having?.cond?.exprs?.length) {
    if (checker(subQuery.having)) {
      subQuery.having = filterExpressions(subQuery.having, checker);
    } else {
      subQuery.having = undefined;
    }
  } else if (subQuery.having) {
    subQuery.having = {
      ...subQuery.having,
    };
  }
  if (subQuery.where?.cond?.exprs?.length) {
    if (checker(subQuery.where)) {
      subQuery.where = filterExpressions(subQuery.where, checker);
    } else {
      subQuery.where = undefined;
    }
  } else if (subQuery.where) {
    subQuery.where = {
      ...subQuery.where,
    };
  }

  return <V1Subquery>{
    dimension: subQuery.dimension,
    measures: [...(subQuery.measures ?? [])],
    where: subQuery.where,
    having: subQuery.having,
  };
}

export function copyFilterExpression(expr: V1Expression) {
  return filterExpressions(expr, () => true) ?? createAndExpression([]);
}

export function filterIdentifiers(
  expr: V1Expression,
  cb: (e: V1Expression, ident: string) => boolean,
) {
  return filterExpressions(expr, (e) => {
    if (e.subquery?.dimension) {
      return cb(e, e.subquery.dimension);
    }
    const ident = e.cond?.exprs?.[0].ident;
    if (ident === undefined) {
      return true;
    }
    return cb(e, ident);
  });
}

export function getValueIndexInExpression(
  expr: V1Expression | undefined,
  value: string,
) {
  return expr?.cond?.exprs?.findIndex((e, i) => i > 0 && e.val === value) ?? -1;
}

export function getValuesInExpression(expr?: V1Expression): any[] {
  return expr ? (expr.cond?.exprs?.slice(1).map((e) => e.val) ?? []) : [];
}

export const matchExpressionByName = (e: V1Expression, name: string) => {
  return e.cond?.exprs?.[0].ident === name;
};

export const sanitiseExpression = (
  where: V1Expression | undefined,
  having: V1Expression | undefined,
) => {
  if (!having) {
    if (!where?.cond?.exprs?.length) return undefined;
    return where;
  }
  if (!where?.cond?.exprs?.length) {
    where = having;
  } else {
    // make sure to create a copy and not update the original "where" filter
    where = createAndExpression([
      ...where.cond.exprs,
      ...(having.cond?.exprs ?? []),
    ]);
  }
  if (!where?.cond?.exprs?.length) return undefined;
  return where;
};

// Check if the operation is unspecified at any level of the condition.
function isOperationUnspecified(cond: V1Condition): boolean {
  if (cond.op === V1Operation.OPERATION_UNSPECIFIED || cond.op === undefined) {
    return true;
  }
  // Check nested conditions
  return (
    cond.exprs?.some(
      (expr) => expr.cond && isOperationUnspecified(expr.cond),
    ) ?? false
  );
}

// Check if the val is defined and non-empty at any level of the nested expressions.
function isValDefinedAndNonEmpty(expr: V1Expression): boolean {
  if (expr.val !== undefined && expr.val !== "") {
    return true; // val is defined and non-empty
  }
  // If there is a nested condition, check if any nested expression has a defined and non-empty val
  return (
    expr.cond?.exprs?.some((nestedExpr) =>
      isValDefinedAndNonEmpty(nestedExpr),
    ) ?? false
  );
}

export function isExpressionIncomplete(expression: V1Expression): boolean {
  // Check the top-level expression's operation
  if (expression.cond && isOperationUnspecified(expression.cond)) {
    return true; // The top-level operation is unspecified, thus incomplete
  }

  // If there's no val at the top level, check nested expressions
  if (!isValDefinedAndNonEmpty(expression)) {
    return true; // No defined and non-empty val found in any expressions, thus incomplete
  }

  // If the operation is specified and a defined, non-empty val is found, the expression is complete
  return false;
}

export function isAndOrExpression(expression: V1Expression | undefined) {
  return (
    expression?.cond?.op &&
    (expression.cond.op === V1Operation.OPERATION_AND ||
      expression.cond.op === V1Operation.OPERATION_OR)
  );
}

export function removeWrapperAndOrExpression(
  expression: V1Expression | undefined,
) {
  if (
    !expression ||
    !isAndOrExpression(expression) ||
    (expression.cond?.exprs?.length && expression.cond.exprs.length > 1)
  )
    return expression;
  return expression.cond?.exprs?.[0];
}

const SupportedOperations = new Set<V1Operation>([
  V1Operation.OPERATION_IN,
  V1Operation.OPERATION_NIN,
  V1Operation.OPERATION_LIKE,
  V1Operation.OPERATION_NLIKE,
]);
export function isExpressionUnsupported(expression: V1Expression) {
  if (
    !expression.cond ||
    !expression.cond.exprs ||
    expression.cond?.op !== V1Operation.OPERATION_AND
  ) {
    return true;
  }

  for (const expr of expression.cond.exprs) {
    if (!expr.cond?.op || !SupportedOperations.has(expr.cond.op)) return true;

    const subqueryExpr = expr.cond?.exprs?.[1];
    if (
      subqueryExpr?.subquery &&
      isSubqueryExpressionUnsupported(subqueryExpr.subquery)
    ) {
      return true;
    }
  }

  return false;
}

export function isSubqueryExpressionUnsupported(subquery: V1Subquery) {
  // While all the types support multiple measure filters per dimension our UI doesn't allow this right now.
  // So unwrap while trying to validate a measure filter.
  const unwrappedHavingFilter = removeWrapperAndOrExpression(subquery.having);
  return (
    unwrappedHavingFilter &&
    isAndOrExpression(unwrappedHavingFilter) &&
    !isBetweenExpression(unwrappedHavingFilter)
  );
}

export function buildValidMetricsViewFilter(
  filter: V1Expression,
  dtf: DimensionThresholdFilter[],
  dimensions: MetricsViewSpecDimension[],
  measures: MetricsViewSpecMeasure[],
) {
  const whereFilter =
    filterIdentifiers(filter, (e, ident) => {
      const dim = dimensions?.find((d) => d.name === ident);
      // ignore if dimension is not present anymore
      if (!dim) return false;
      return true;
    }) ?? createAndExpression([]);

  const dimensionThresholdFilter = dtf.filter((f) => {
    const dim = dimensions?.find((d) => d.name === f.name);
    if (!dim) return false;

    const hasValidMeasures = f.filters.every((filter) => {
      const measure = measures?.find((m) => m.name === filter.measure);
      return !!measure;
    });
    return hasValidMeasures;
  });

  return sanitiseExpression(
    mergeDimensionAndMeasureFilters(whereFilter, dimensionThresholdFilter),
    undefined,
  );
}

export function wrapNonJoinerExpression(expr: V1Expression): V1Expression {
  if (isAndOrExpression(expr)) return expr;
  return createAndExpression([expr]);
}

export function maybeConvertEqualityToInExpressions(expr: V1Expression) {
  if (
    !expr.cond?.op ||
    !expr.cond.exprs ||
    (expr.cond.op !== V1Operation.OPERATION_EQ &&
      expr.cond.op !== V1Operation.OPERATION_NEQ)
  ) {
    return expr;
  }

  const ident = expr.cond.exprs[0]?.ident;
  if (!ident) return expr;

  const valExprs = expr.cond.exprs.slice(1);
  if (valExprs.some((ve) => !("val" in ve))) return expr;

  const vals = valExprs.map((e) => e.val);

  return createInExpression(
    ident,
    vals,
    expr.cond.op === V1Operation.OPERATION_NEQ,
  );
}

/**
 * Flattens the value part of in/nin expression. Eg,
 * [ {ident}, {val: [a,b]} ] ==> [ {ident}, {val: a}, {val: b} ]
 * This is needed to correct possibly correct filter sent by LLM.
 */
export function flattenInExpressionValues(expr: V1Expression) {
  const notInExpr =
    !expr.cond?.op ||
    (expr.cond.op !== V1Operation.OPERATION_IN &&
      expr.cond.op !== V1Operation.OPERATION_NIN);
  if (notInExpr) return expr;
  const ident = expr.cond!.exprs?.[0]?.ident;
  const vals = expr.cond!.exprs?.slice(1).flatMap((e) => e.val);
  if (!ident || !vals) return expr;
  return createInExpression(
    ident,
    vals,
    expr.cond!.op === V1Operation.OPERATION_NIN,
  );
}
