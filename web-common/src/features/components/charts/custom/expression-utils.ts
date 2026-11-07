import type {
  V1Expression,
  V1Operation,
  V1Subquery,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

// TODO: Update resolver API to accept V1Expression directly

/**
 * Maps proto V1Operation enum strings to the short operator strings
 * expected by Go's metricsview.Expression mapstructure tags.
 * See: runtime/metricsview/query_expression_pb.go NewExpressionFromProto
 */
const OperatorMap: Record<string, string> = {
  OPERATION_EQ: "eq",
  OPERATION_NEQ: "neq",
  OPERATION_LT: "lt",
  OPERATION_LTE: "lte",
  OPERATION_GT: "gt",
  OPERATION_GTE: "gte",
  OPERATION_OR: "or",
  OPERATION_AND: "and",
  OPERATION_IN: "in",
  OPERATION_NIN: "nin",
  OPERATION_LIKE: "ilike",
  OPERATION_NLIKE: "nilike",
  OPERATION_CAST: "cast",
};

/**
 * Converts a V1Expression (proto-generated type) to the plain object format
 * expected by Go's metricsview.Expression when passed through resolverProperties
 * (google.protobuf.Struct / mapstructure).
 *
 * Key differences:
 * - Proto uses `ident` for identifiers; Go mapstructure expects `name`
 * - Proto uses `OPERATION_*` enum strings; Go expects short strings like "and", "eq"
 */
export function convertV1ExpressionToMapstructure(
  expr: V1Expression,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  if (expr.ident !== undefined) {
    result.name = expr.ident;
  }

  if (expr.val !== undefined) {
    result.val = expr.val;
  }

  if (expr.cond) {
    result.cond = {
      op: expr.cond.op ? (OperatorMap[expr.cond.op] ?? expr.cond.op) : "",
      exprs: expr.cond.exprs?.map(convertV1ExpressionToMapstructure) ?? [],
    };
  }

  if (expr.subquery) {
    result.subquery = expr.subquery;
  }

  return result;
}

const ReverseOperatorMap: Record<string, string> = Object.fromEntries(
  Object.entries(OperatorMap).map(([proto, short]) => [short, proto]),
);

/**
 * The inverse of convertV1ExpressionToMapstructure: reads an expression stored in resolverProperties back as a V1Expression.
 */
export function convertMapstructureToV1Expression(
  expr: Record<string, unknown>,
): V1Expression {
  const result: V1Expression = {};

  if (expr.name !== undefined) {
    result.ident = expr.name as string;
  }

  if (expr.val !== undefined) {
    result.val = expr.val;
  }

  const cond = expr.cond as
    | { op?: string; exprs?: Record<string, unknown>[] }
    | undefined;
  if (cond) {
    result.cond = {
      op: (ReverseOperatorMap[cond.op ?? ""] ?? cond.op) as V1Operation,
      exprs: cond.exprs?.map(convertMapstructureToV1Expression) ?? [],
    };
  }

  if (expr.subquery) {
    result.subquery = expr.subquery as V1Subquery;
  }

  return result;
}

/**
 * The metrics_sql resolver properties that apply a dashboard's filters to a query:
 * the where filter (when it has conditions) and the time range.
 */
export function metricsSQLFilterProperties(
  whereFilter: V1Expression | undefined,
  timeRange: V1TimeRange | undefined,
): Record<string, unknown> {
  return {
    ...(whereFilter?.cond?.exprs?.length
      ? { additional_where: convertV1ExpressionToMapstructure(whereFilter) }
      : {}),
    ...(timeRange ? { additional_time_range: timeRange } : {}),
  };
}
