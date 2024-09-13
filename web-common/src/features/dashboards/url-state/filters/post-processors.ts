import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Operation } from "@rilldata/web-common/runtime-client";

const BinaryOperationMap = {
  "=": V1Operation.OPERATION_EQ,
  "!=": V1Operation.OPERATION_NEQ,
  ">": V1Operation.OPERATION_GT,
  ">=": V1Operation.OPERATION_GTE,
  "<": V1Operation.OPERATION_LT,
  "<=": V1Operation.OPERATION_LTE,
};

export const binaryPostprocessor = ([left, _1, op, _2, right]) =>
  createBinaryExpression(left, BinaryOperationMap[op], right);

export const inPostprocessor = ([column, _1, op, _2, _3, values]) =>
  createInExpression(column, values, op === "NIN");

export const havingPostprocessor = ([column, _1, _2, _3, _4, expr]) =>
  createSubQueryExpression(column, getAllIdentifiers(expr), expr);

export const andOrPostprocessor = ([left, right]) => {
  const op = left[0][2].toUpperCase();
  const exprs = [...left.map((t) => t[0]), right];
  if (op === "AND") return createAndExpression(exprs);
  return createOrExpression(exprs);
};
