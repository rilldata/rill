import {
  flattenInExpressionValues,
  maybeConvertEqualityToInExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

const OperationMap: Record<string, V1Operation> = {
  "": V1Operation.OPERATION_UNSPECIFIED,
  eq: V1Operation.OPERATION_EQ,
  neq: V1Operation.OPERATION_NEQ,
  lt: V1Operation.OPERATION_LT,
  lte: V1Operation.OPERATION_LTE,
  gt: V1Operation.OPERATION_GT,
  gte: V1Operation.OPERATION_GTE,
  in: V1Operation.OPERATION_IN,
  nin: V1Operation.OPERATION_NIN,
  ilike: V1Operation.OPERATION_LIKE,
  nilike: V1Operation.OPERATION_NLIKE,
  or: V1Operation.OPERATION_OR,
  and: V1Operation.OPERATION_AND,
};

export function mapResolverExpressionToV1Expression(
  expr: Expression | undefined,
): V1Expression | undefined {
  if (!expr) return undefined;

  if (expr.name) {
    return { ident: expr.name };
  }

  // Check presence rather than truthiness: 0, false and "" are valid filter values.
  if (expr.val !== undefined) {
    return { val: expr.val };
  }

  if (expr.cond) {
    const condExpr = maybeConvertEqualityToInExpressions({
      cond: {
        op: OperationMap[expr.cond.op] || V1Operation.OPERATION_UNSPECIFIED,
        exprs: expr.cond.exprs
          ?.map(mapResolverExpressionToV1Expression)
          .filter((expr) => expr !== undefined),
      },
    });
    return flattenInExpressionValues(condExpr);
  }

  if (expr.subquery) {
    return {
      subquery: {
        dimension: expr.subquery.dimension.name,
        measures: expr.subquery.measures.map((m) => m.name),
        where: mapResolverExpressionToV1Expression(expr.subquery.where),
        having: mapResolverExpressionToV1Expression(expr.subquery.having),
      },
    };
  }

  return {};
}
