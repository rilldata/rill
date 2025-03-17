import { BinaryOperationReverseMap } from "@rilldata/web-common/features/dashboards/url-state/filters/post-processors";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import grammar from "./expression.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function convertFilterParamToExpression(filter: string) {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(filter);
  return parser.results[0] as V1Expression;
}

const NonStandardName = /^[a-zA-Z][a-zA-Z0-9_]*$/;
export function convertExpressionToFilterParam(
  expr: V1Expression,
  depth = 0,
): string {
  if (!expr) return "";

  if ("val" in expr) {
    return escapeValue(expr.val);
  }

  if (expr.ident) return escapeColumnName(expr.ident);

  switch (expr.cond?.op) {
    case V1Operation.OPERATION_AND:
    case V1Operation.OPERATION_OR:
      return convertJoinerExpressionToFilterParam(expr, depth);

    case V1Operation.OPERATION_IN:
    case V1Operation.OPERATION_NIN:
      return convertInExpressionToFilterParam(expr, depth);

    default:
      return convertBinaryExpressionToFilterParam(expr, depth);
  }
}

export function stripParserError(err: Error) {
  return err.message.substring(
    0,
    err.message.indexOf("Instead, I was expecting") - 1,
  );
}

function convertJoinerExpressionToFilterParam(
  expr: V1Expression,
  depth: number,
) {
  const joiner = expr.cond?.op === V1Operation.OPERATION_AND ? " AND " : " OR ";

  const parts = expr.cond?.exprs
    ?.map((e) => convertExpressionToFilterParam(e, depth + 1))
    .filter(Boolean);
  if (!parts?.length) return "";
  const exprParam = parts.join(joiner);

  if (depth === 0) {
    return exprParam;
  }
  return `(${exprParam})`;
}

function convertInExpressionToFilterParam(expr: V1Expression, depth: number) {
  if (!expr.cond?.exprs?.length) return "";
  let joiner = expr.cond?.op === V1Operation.OPERATION_IN ? "IN" : "NIN";
  const isMatchList = !!(expr as any).isMatchList;
  if (isMatchList) {
    joiner =
      expr.cond?.op === V1Operation.OPERATION_IN ? "IN LIST" : "NOT IN LIST";
  }

  const column = expr.cond.exprs[0]?.ident;
  if (!column) return "";
  const safeColumn = escapeColumnName(column);

  if (expr.cond.exprs[1]?.subquery?.having) {
    // TODO: support `NIN <subquery>`
    const having = convertExpressionToFilterParam(
      expr.cond.exprs[1]?.subquery?.having,
      0,
    );
    if (having) return `${safeColumn} having (${having})`;
  }

  if (expr.cond.exprs.length > 1) {
    const vals = expr.cond.exprs
      .slice(1)
      .map((e) => convertExpressionToFilterParam(e, depth + 1));
    return `${safeColumn} ${joiner} (${vals.join(",")})`;
  }

  return "";
}

function convertBinaryExpressionToFilterParam(
  expr: V1Expression,
  depth: number,
) {
  if (!expr.cond?.op || !(expr.cond?.op in BinaryOperationReverseMap))
    return "";
  const op = BinaryOperationReverseMap[expr.cond.op];
  if (!expr.cond?.exprs?.length) return "";
  const left = convertExpressionToFilterParam(expr.cond.exprs[0], depth + 1);
  const right = convertExpressionToFilterParam(expr.cond.exprs[1], depth + 1);
  if (!left || !right) return "";

  return `${left} ${op?.toUpperCase()} ${right}`;
}

function escapeColumnName(columnName: string) {
  // if name doesnt have any special chars do not surround it by quotes.
  // this makes the url more readable
  if (NonStandardName.test(columnName)) return columnName;
  const escapedColumnName = columnName
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"');
  return `"${escapedColumnName}"`;
}

function escapeValue(value: unknown) {
  switch (typeof value) {
    case "string":
      return escapeStringValue(value);

    case "object":
      if (!value) return "null";
      if (Array.isArray(value)) {
        return `[${value.map(escapeValue).join(",")}]`;
      }
      return `{${Object.keys(value)
        .map((k) => `'${k}':${escapeValue(value[k])}`)
        .join(",")}}`;
  }

  return value + "";
}

function escapeStringValue(value: string) {
  const escapedValue = value
    // TODO: this was a CodeQL suggestion. could this cause conflicts in values?
    .replace(/\\/g, "\\\\")
    .replace(/'/g, "\\'")
    .replace(/\n/g, "\\n");
  return `'${escapedValue}'`;
}
