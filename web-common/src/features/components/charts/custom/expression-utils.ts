import type { V1Expression } from "@rilldata/web-common/runtime-client";

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
