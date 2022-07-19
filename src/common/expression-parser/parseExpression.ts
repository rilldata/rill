import { Expr, ExprBinary, ExprCall, parse } from "pgsql-ast-parser";
import type { InvalidAggregate } from "$common/expression-parser/validateAggregate";
import { validateAggregate } from "$common/expression-parser/validateAggregate";
import type { ProfileColumn } from "$lib/types";

export interface ParseExpressionError {
  message?: string;
  location?: { start: number; end: number };
  unexpectedToken?: string;
  disallowedSyntax?: string;
  missingColumns?: Array<string>;
  missingFrom?: string;
  invalidAggregates?: Array<InvalidAggregate>;
}
export interface ParsedExpression {
  expression: string;
  isValid: boolean;
  columns: Array<string>;
  error?: ParseExpressionError;
}

export function parseExpression(
  expression: string,
  profileColumns: Array<ProfileColumn>
): ParsedExpression {
  try {
    const expr = parse(expression, {
      entry: "expr",
      locationTracking: true,
    }) as unknown as Expr;
    const parsedExpression: ParsedExpression = {
      expression,
      isValid: true,
      columns: [],
    };
    buildParsedExpression(expr, parsedExpression, profileColumns);
    return parsedExpression;
  } catch (err) {
    return {
      expression,
      isValid: false,
      columns: [],
      error: {
        message: err.message,
        location: err.token?._location ?? {
          start: expression.length - 1,
          end: expression.length - 1,
        },
        unexpectedToken: err.token?.text,
      },
    };
  }
}

function buildParsedExpression(
  expr: Expr,
  parsedExpression: ParsedExpression,
  profileColumns: Array<ProfileColumn>
) {
  switch (expr.type) {
    case "call":
      buildParsedExpressionFromCall(expr, parsedExpression, profileColumns);
      break;

    case "binary":
      buildParsedExpressionFromBinary(expr, parsedExpression, profileColumns);
      break;

    case "cast":
      buildParsedExpressionInner(
        expr.operand,
        parsedExpression,
        profileColumns
      );
      break;

    case "integer":
    case "numeric":
    case "boolean":
    case "string":
      break;

    default:
      parsedExpression.error = {
        message: "disallowed syntax",
        disallowedSyntax: expr.type,
        location: expr._location,
      };
      parsedExpression.isValid = false;
  }
}

function buildParsedExpressionInner(
  expr: Expr,
  parsedExpression: ParsedExpression,
  profileColumns: Array<ProfileColumn>
) {
  if (expr.type === "ref") {
    parsedExpression.columns.push(expr.name);
    return expr.name;
    // TODO: arg.table
  } else {
    buildParsedExpression(expr, parsedExpression, profileColumns);
    return "";
  }
}

function buildParsedExpressionFromCall(
  expr: ExprCall,
  parsedExpression: ParsedExpression,
  profileColumns: Array<ProfileColumn>
) {
  const args = new Array<string>();
  for (const arg of expr.args) {
    args.push(
      buildParsedExpressionInner(arg, parsedExpression, profileColumns)
    );
  }

  const aggregateError = validateAggregate(expr, args, profileColumns);
  if (aggregateError) {
    parsedExpression.isValid = false;
    parsedExpression.error ??= {};
    parsedExpression.error.invalidAggregates ??= [];
    parsedExpression.error.invalidAggregates.push(aggregateError);
  }
}

function buildParsedExpressionFromBinary(
  expr: ExprBinary,
  parsedExpression: ParsedExpression,
  profileColumns: Array<ProfileColumn>
) {
  buildParsedExpression(expr.left, parsedExpression, profileColumns);
  buildParsedExpression(expr.right, parsedExpression, profileColumns);
}
