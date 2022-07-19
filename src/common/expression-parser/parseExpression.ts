import { Expr, ExprBinary, ExprCall, parse } from "pgsql-ast-parser";

export interface ParseExpressionError {
  message?: string;
  location?: { start: number; end: number };
  unexpectedToken?: string;
  disallowedSyntax?: string;
  missingColumns?: Array<string>;
  missingFrom?: string;
}
export interface ParsedExpression {
  expression: string;
  isValid: boolean;
  columns: Array<string>;
  error?: ParseExpressionError;
}

export function parseExpression(expression: string): ParsedExpression {
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
    buildParsedExpression(expr, parsedExpression);
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

function buildParsedExpression(expr: Expr, parsedExpression: ParsedExpression) {
  switch (expr.type) {
    case "call":
      buildParsedExpressionFromCall(expr, parsedExpression);
      break;

    case "binary":
      buildParsedExpressionFromBinary(expr, parsedExpression);
      break;

    case "cast":
      buildParsedExpressionInner(expr.operand, parsedExpression);
      break;

    case "integer":
    case "numeric":
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
  parsedExpression: ParsedExpression
) {
  if (expr.type === "ref") {
    parsedExpression.columns.push(expr.name);
    // TODO: arg.table
  } else {
    buildParsedExpression(expr, parsedExpression);
  }
}

function buildParsedExpressionFromCall(
  expr: ExprCall,
  parsedExpression: ParsedExpression
) {
  for (const arg of expr.args) {
    buildParsedExpressionInner(arg, parsedExpression);
  }
}

function buildParsedExpressionFromBinary(
  expr: ExprBinary,
  parsedExpression: ParsedExpression
) {
  buildParsedExpression(expr.left, parsedExpression);
  buildParsedExpression(expr.right, parsedExpression);
}
