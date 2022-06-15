import { Expr, ExprBinary, ExprCall, parse } from "pgsql-ast-parser";
import type { SelectedColumn, Statement } from "pgsql-ast-parser";

export interface ParsedExpression {
  isValid: boolean;
  columns: Array<string>;
}

export function parseQuery(query: string): SelectedColumn[] {
  try {
    return getColumns(parse(query, { locationTracking: true })[0]);
  } catch (err) {
    console.error(err);
  }
  return [];
}

function getColumns(statement: Statement): SelectedColumn[] {
  if (statement.type === "select") {
    return statement.columns;
  } else if (statement.type === "with" || statement.type === "with recursive") {
    return getColumns(statement.in);
  }
  return [];
}

export function parseExpression(expression: string): ParsedExpression {
  try {
    const expr = parse(expression, {
      entry: "expr",
      locationTracking: true,
    }) as unknown as Expr;
    const parsedExpression: ParsedExpression = {
      isValid: true,
      columns: [],
    };
    buildParsedExpression(expr, parsedExpression);
    return parsedExpression;
  } catch (err) {
    return {
      isValid: false,
      columns: [],
    };
  }
}

function buildParsedExpression(expr: Expr, parsedExpression: ParsedExpression) {
  if (expr.type === "call") {
    buildParsedExpressionFromCall(expr, parsedExpression);
  } else if (expr.type === "binary") {
    buildParsedExpressionFromBinary(expr, parsedExpression);
  } else {
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
  buildParsedExpressionInner(expr.left, parsedExpression);
  buildParsedExpressionInner(expr.right, parsedExpression);
}
