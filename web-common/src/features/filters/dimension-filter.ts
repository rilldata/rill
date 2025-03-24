import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export enum DimensionFilterMode {
  Select = "Select",
  Contains = "Contains",
  InList = "InList",
}

export type DimensionFilterEntry = {
  name: string;
  mode: DimensionFilterMode;
  exclude: boolean;

  values?: unknown[];
  inputText?: string;
};

export type DimensionFilterDisplayEntry = DimensionFilterEntry & {
  label: string;
};

export function mapExprToDimensionFilter(
  expr: V1Expression | undefined,
): DimensionFilterEntry | undefined {
  if (!expr) return undefined;

  const name = expr.cond?.exprs?.[0].ident;
  if (!name) return undefined;

  const isExclude =
    expr.cond?.op === V1Operation.OPERATION_NIN ||
    expr.cond?.op === V1Operation.OPERATION_NLIKE ||
    expr.cond?.op === V1Operation.OPERATION_NEQ;

  if (
    expr.cond?.op === V1Operation.OPERATION_IN ||
    expr.cond?.op === V1Operation.OPERATION_NIN
  ) {
    return {
      name,
      mode: DimensionFilterMode.Select,
      exclude: isExclude,
      values: expr.cond?.exprs?.splice(1).map((e) => e.val) ?? [],
    };
  } else if (
    expr.cond?.op === V1Operation.OPERATION_LIKE ||
    expr.cond?.op === V1Operation.OPERATION_NLIKE
  ) {
    return {
      name,
      mode: DimensionFilterMode.Contains,
      exclude: isExclude,
      inputText: (expr.cond?.exprs?.[1]?.val as string) ?? "",
    };
  }

  return undefined;
}

export function mapDimensionFilterToExpr(
  dimensionFilter: DimensionFilterEntry,
): V1Expression {
  switch (dimensionFilter.mode) {
    case DimensionFilterMode.Select:
    case DimensionFilterMode.InList:
      return createInExpression(
        dimensionFilter.name,
        dimensionFilter.values ?? [],
        dimensionFilter.exclude,
      );

    case DimensionFilterMode.Contains:
      return createLikeExpression(
        dimensionFilter.name,
        dimensionFilter.inputText ?? "",
        dimensionFilter.exclude,
      );
  }
}
