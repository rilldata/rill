import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";
import grammar from "./expression.js";
import nearley from "nearley";

export function convertFilterParamToExpression(filter: string) {
  const parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
  parser.feed(filter);
  return convertToExpression(parser.results[0]);
}

const BinaryOperationMap = {
  "=": V1Operation.OPERATION_EQ,
  "!=": V1Operation.OPERATION_NEQ,
  ">": V1Operation.OPERATION_GT,
  ">=": V1Operation.OPERATION_GTE,
  "<": V1Operation.OPERATION_LT,
  "<=": V1Operation.OPERATION_LTE,
};
function convertToExpression(parsed: any) {
  let subExpr: V1Expression | undefined;
  switch (typeof parsed) {
    case "object":
      switch (parsed?.[0]) {
        case "IN":
          return createInExpression(parsed[1], parsed[2]);
        case "NIN":
          return createInExpression(parsed[1], parsed[2], true);

        case "=":
        case "!=":
        case ">":
        case ">=":
        case "<":
        case "<=":
          return createBinaryExpression(
            parsed[1],
            BinaryOperationMap[parsed[0]],
            parsed[2],
          );

        case "AND":
          return createAndExpression(parsed.slice(1).map(convertToExpression));
        case "OR":
          return createOrExpression(parsed.slice(1).map(convertToExpression));

        case "HAVING":
          subExpr = convertToExpression(parsed[2]);
          return createSubQueryExpression(
            parsed[1],
            getAllIdentifiers(subExpr),
            subExpr,
          );
      }
      break;

    default:
      return parsed;
  }

  return undefined;
}
