import { forEachIdentifier } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { BinaryOperationReverseMap } from "@rilldata/web-common/features/dashboards/url-state/filters/post-processors";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import nearley from "nearley";
import { isNonStandardIdentifier } from "../../../entity-management/name-utils";
import grammar from "./expression.cjs";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function convertFilterParamToExpression(filter: string): {
  expr: V1Expression | undefined;
  dimensionsWithInlistFilter: string[];
} {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(filter);
  const expr = parser.results[0] as V1Expression;
  const dimensionsWithInlistFilter: string[] = [];

  if (!expr) {
    return { expr: undefined, dimensionsWithInlistFilter };
  }

  forEachIdentifier(expr, (e, ident) => {
    if ((e as any).isInListMode) {
      dimensionsWithInlistFilter.push(ident);
      delete (e as any).isInListMode;
    }
  });

  return { expr, dimensionsWithInlistFilter };
}

export function convertExpressionToFilterParam(
  expr: V1Expression,
  dimensionsWithInlistFilter: string[] = [],
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
      return convertJoinerExpressionToFilterParam(
        expr,
        dimensionsWithInlistFilter,
        depth,
      );

    case V1Operation.OPERATION_IN:
    case V1Operation.OPERATION_NIN:
      return convertInExpressionToFilterParam(
        expr,
        dimensionsWithInlistFilter,
        depth,
      );

    default:
      return convertBinaryExpressionToFilterParam(
        expr,
        dimensionsWithInlistFilter,
        depth,
      );
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
  dimensionsWithInlistFilter: string[],
  depth: number,
) {
  const joiner = expr.cond?.op === V1Operation.OPERATION_AND ? " AND " : " OR ";

  const parts = expr.cond?.exprs
    ?.map((e) =>
      convertExpressionToFilterParam(e, dimensionsWithInlistFilter, depth + 1),
    )
    .filter(Boolean);
  if (!parts?.length) return "";
  const exprParam = parts.join(joiner);

  if (depth === 0) {
    return exprParam;
  }
  return `(${exprParam})`;
}

function convertInExpressionToFilterParam(
  expr: V1Expression,
  dimensionsWithInlistFilter: string[],
  depth: number,
) {
  if (!expr.cond?.exprs?.length) return "";
  const column = expr.cond.exprs[0]?.ident;
  if (!column) return "";
  const safeColumn = escapeColumnName(column);

  let joiner = expr.cond?.op === V1Operation.OPERATION_IN ? "IN" : "NIN";
  const isInListFilter = dimensionsWithInlistFilter.includes(column);
  if (isInListFilter) {
    joiner =
      expr.cond?.op === V1Operation.OPERATION_IN ? "IN LIST" : "NOT IN LIST";
  }

  if (expr.cond.exprs[1]?.subquery?.having) {
    // TODO: support `NIN <subquery>`
    const having = convertExpressionToFilterParam(
      expr.cond.exprs[1]?.subquery?.having,
      dimensionsWithInlistFilter,
      0,
    );
    if (having) return `${safeColumn} having (${having})`;
  }

  if (expr.cond.exprs.length > 1) {
    const vals = expr.cond.exprs
      .slice(1)
      .map((e) =>
        convertExpressionToFilterParam(
          e,
          dimensionsWithInlistFilter,
          depth + 1,
        ),
      );
    return `${safeColumn} ${joiner} (${vals.join(",")})`;
  }

  return "";
}

function convertBinaryExpressionToFilterParam(
  expr: V1Expression,
  dimensionsWithInlistFilter: string[],
  depth: number,
) {
  if (!expr.cond?.op || !(expr.cond?.op in BinaryOperationReverseMap))
    return "";
  const op = BinaryOperationReverseMap[expr.cond.op];
  if (!expr.cond?.exprs?.length) return "";
  const left = convertExpressionToFilterParam(
    expr.cond.exprs[0],
    dimensionsWithInlistFilter,
    depth + 1,
  );
  const right = convertExpressionToFilterParam(
    expr.cond.exprs[1],
    dimensionsWithInlistFilter,
    depth + 1,
  );
  if (!left || !right) return "";

  return `${left} ${op?.toUpperCase()} ${right}`;
}

function escapeColumnName(columnName: string) {
  if (isNonStandardIdentifier(columnName)) {
    const escapedColumnName = columnName
      .replace(/\\/g, "\\\\")
      .replace(/"/g, '\\"');
    return `"${escapedColumnName}"`;
  }

  // If name doesn't have any special characters, do not surround it by quotes.
  return columnName;
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
