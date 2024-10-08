import { BinaryOperationReverseMap } from "@rilldata/web-common/features/dashboards/url-state/filters/post-processors";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";
import grammar from "./expression.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function convertFilterParamToExpression(filter: string) {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(filter);
  return parser.results[0] as V1Expression;
}

export function convertExpressionToFilterParam(expr: V1Expression): string {
  if (!expr) return "";

  if (expr.val) {
    if (typeof expr.val === "string") {
      return `'${expr.val}'`;
    }
    return expr.val + "";
  }

  if (expr.ident) return expr.ident;

  switch (expr.cond?.op) {
    case V1Operation.OPERATION_AND:
    case V1Operation.OPERATION_OR:
      return convertJoinerExpressionToFilterParam(expr);

    case V1Operation.OPERATION_IN:
    case V1Operation.OPERATION_NIN:
      return convertInExpressionToFilterParam(expr);

    default:
      return convertBinaryExpressionToFilterParam(expr);
  }
}

export function stripParserError(err: Error) {
  return err.message.substring(
    0,
    err.message.indexOf("Instead, I was expecting") - 1,
  );
}

function convertJoinerExpressionToFilterParam(expr: V1Expression) {
  const joiner = expr.cond?.op === V1Operation.OPERATION_AND ? "AND" : "OR";

  const parts = expr.cond?.exprs
    ?.map(convertExpressionToFilterParam)
    .filter(Boolean);
  if (!parts?.length) return "";

  return `(${parts.join(joiner)})`;
}

function convertInExpressionToFilterParam(expr: V1Expression) {
  if (!expr.cond?.exprs?.length) return "";
  const joiner = expr.cond?.op === V1Operation.OPERATION_IN ? "IN" : "NIN";

  const column = expr.cond.exprs[0]?.ident;
  if (!column) return "";

  if (expr.cond.exprs[1]?.subquery?.having) {
    // TODO: support `NIN <subquery>`
    const having = convertExpressionToFilterParam(
      expr.cond.exprs[1]?.subquery?.having,
    );
    if (having) return `${column} having (${having})`;
  }

  if (expr.cond.exprs.length > 1) {
    const vals = expr.cond.exprs
      .slice(1)
      .map(convertExpressionToFilterParam)
      .filter(Boolean);
    return `${column} ${joiner} (${vals.join(",")})`;
  }

  return "";
}

function convertBinaryExpressionToFilterParam(expr: V1Expression) {
  if (!expr.cond?.op || !(expr.cond?.op in BinaryOperationReverseMap))
    return "";
  const op = BinaryOperationReverseMap[expr.cond.op];
  if (!expr.cond?.exprs?.length) return "";
  const left = convertExpressionToFilterParam(expr.cond.exprs[0]);
  const right = convertExpressionToFilterParam(expr.cond.exprs[1]);
  if (!left || !right) return "";

  return `${left} ${op} ${right}`;
}
