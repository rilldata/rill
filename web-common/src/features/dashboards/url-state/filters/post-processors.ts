import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";

import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { V1Operation } from "@rilldata/web-common/runtime-client";

const BinaryOperationMap: Record<string, V1Operation> = {
  eq: V1Operation.OPERATION_EQ,
  neq: V1Operation.OPERATION_NEQ,
  gt: V1Operation.OPERATION_GT,
  gte: V1Operation.OPERATION_GTE,
  lt: V1Operation.OPERATION_LT,
  lte: V1Operation.OPERATION_LTE,
  like: V1Operation.OPERATION_LIKE,
  nlike: V1Operation.OPERATION_NLIKE,
  "not like": V1Operation.OPERATION_NLIKE,
};
export const BinaryOperationReverseMap = reverseMap(BinaryOperationMap);
BinaryOperationReverseMap[V1Operation.OPERATION_NLIKE] = "not like";

export const binaryPostprocessor = ([left, _1, op, _2, right]) =>
  createBinaryExpression(left, BinaryOperationMap[op.toLowerCase()], right);

export const inPostprocessor = ([column, _1, op, _2, _3, values]: [
  string,
  unknown,
  string,
  unknown,
  unknown,
  any[],
]) => {
  const lowerCaseOperator = op.toLowerCase();
  const isInclude =
    lowerCaseOperator === "in" || lowerCaseOperator === "in list";
  const expr = createInExpression(column, values, !isInclude);
  const isInListMode =
    lowerCaseOperator === "in list" || lowerCaseOperator === "not in list";
  if (isInListMode) {
    (expr as any).isInListMode = isInListMode;
  }
  return expr;
};

export const havingPostprocessor = ([column, _1, _2, _3, _4, expr]) =>
  createSubQueryExpression(column, getAllIdentifiers(expr), expr);

export const andOrPostprocessor = ([left, right]) => {
  const op = left[0][2].toUpperCase();
  const exprs = [...left.map((t) => t[0]), right];
  if (op === "AND") return createAndExpression(exprs);
  return createOrExpression(exprs);
};

export const objectPostprocessor = ([
  _1,
  _2,
  keyValue,
  otherKeyValuesMatches,
]) => {
  const obj = { ...keyValue };
  otherKeyValuesMatches.forEach(([_1, _2, _3, keyValue]) => {
    Object.assign(obj, keyValue);
  });
  return obj;
};
