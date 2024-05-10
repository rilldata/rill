import {
  MeasureFilterOperation,
  MeasureFilterToProtoOperation,
  ProtoToCompareMeasureFilterOperation,
  ProtoToMeasureFilterOperations,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createBetweenExpression,
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";

export enum MeasureFilterComparisonType {
  None,
  AbsoluteComparison,
  PercentageComparison,
  AbsoluteShareOfTotal,
  PercentageShareOfTotal,
}

export type MeasureFilterEntry = {
  measure: string;
  operation: MeasureFilterOperation;
  comparison: MeasureFilterComparisonType;
  value1: string;
  value2: string;
};

const DeltaAbsoluteSuffix = "__delta_abs";
const DeltaRelativeSuffix = "__delta_rel";
const HasSuffixRegex = /__delta_(rel|abs)/;

export function mapExprToMeasureFilter(
  expr: V1Expression | undefined,
): MeasureFilterEntry | undefined {
  if (!expr) return undefined;

  let value1 = 0;
  let value2: number | undefined;
  let field = "";
  let operation = MeasureFilterOperation.GreaterThan;
  let comparison = MeasureFilterComparisonType.None;

  switch (expr.cond?.op) {
    case V1Operation.OPERATION_OR:
      field = expr.cond.exprs?.[0].cond?.exprs?.[0].ident ?? "";
      if (HasSuffixRegex.test(field)) {
        // handle ChangeBy
        value1 =
          (expr.cond.exprs?.[1].cond?.exprs?.[1].val as number) * 100 ?? 0;
        operation = MeasureFilterOperation.ChangesBy;
        break;
      }
    // eslint-disable-next-line no-fallthrough
    case V1Operation.OPERATION_AND:
      // handle between and not-between
      field = expr.cond.exprs?.[0].cond?.exprs?.[0].ident ?? "";
      value1 = (expr.cond.exprs?.[0].cond?.exprs?.[1].val as number) ?? 0;
      value2 = (expr.cond.exprs?.[1].cond?.exprs?.[1].val as number) ?? 0;
      operation =
        expr.cond?.op === V1Operation.OPERATION_AND
          ? MeasureFilterOperation.Between
          : MeasureFilterOperation.NotBetween;
      break;

    case V1Operation.OPERATION_EQ:
    case V1Operation.OPERATION_NEQ:
    case V1Operation.OPERATION_GT:
    case V1Operation.OPERATION_GTE:
    case V1Operation.OPERATION_LT:
    case V1Operation.OPERATION_LTE:
      field = expr.cond.exprs?.[0].ident ?? "";
      value1 = Math.abs((expr.cond.exprs?.[1].val as number) ?? 0);
      if (field.endsWith(DeltaRelativeSuffix)) {
        // convert decimal to percent
        value1 *= 100;
      }
      operation =
        (HasSuffixRegex.test(field)
          ? ProtoToCompareMeasureFilterOperation[expr.cond?.op]
          : ProtoToMeasureFilterOperations[expr.cond?.op]) ??
        MeasureFilterOperation.GreaterThan;
      break;
  }

  if (field.endsWith(DeltaAbsoluteSuffix)) {
    comparison = MeasureFilterComparisonType.AbsoluteComparison;
  } else if (field.endsWith(DeltaRelativeSuffix)) {
    comparison = MeasureFilterComparisonType.PercentageComparison;
  }

  return {
    measure: field.replace(HasSuffixRegex, ""),
    value1: value1.toString(),
    value2: value2?.toString() ?? "",
    operation,
    comparison,
  };
}

export function mapMeasureFilterToExpr(
  measureFilter: MeasureFilterEntry,
): V1Expression | undefined {
  let value = Number(measureFilter.value1);
  if (Number.isNaN(value)) {
    return undefined;
  }

  if (
    measureFilter.comparison ===
    MeasureFilterComparisonType.PercentageComparison
  ) {
    value /= 100;
  }
  const comparisonSuffix =
    measureFilter.comparison ===
    MeasureFilterComparisonType.PercentageComparison
      ? DeltaRelativeSuffix
      : DeltaAbsoluteSuffix;

  switch (measureFilter.operation) {
    case MeasureFilterOperation.Equals:
    case MeasureFilterOperation.NotEquals:
    case MeasureFilterOperation.GreaterThan:
    case MeasureFilterOperation.GreaterThanOrEquals:
    case MeasureFilterOperation.LessThan:
    case MeasureFilterOperation.LessThanOrEquals:
      if (measureFilter.comparison !== MeasureFilterComparisonType.None)
        return undefined;
      return createBinaryExpression(
        measureFilter.measure,
        MeasureFilterToProtoOperation[measureFilter.operation],
        value,
      );

    case MeasureFilterOperation.Between:
    case MeasureFilterOperation.NotBetween:
      return createBetweenExpression(
        measureFilter.measure,
        value,
        Number(measureFilter.value2 ?? "0"),
        measureFilter.operation === MeasureFilterOperation.NotBetween,
      );

    case MeasureFilterOperation.IncreasesBy:
      // Δ<field> > <value>
      // or
      // Δ%<field> > <value>
      return createBinaryExpression(
        measureFilter.measure + comparisonSuffix,
        V1Operation.OPERATION_GT,
        value,
      );

    case MeasureFilterOperation.DecreasesBy:
      // Δ<field> < -<value>
      // or
      // Δ%<field> < -<value>
      return createBinaryExpression(
        measureFilter.measure + comparisonSuffix,
        V1Operation.OPERATION_LT,
        -Math.abs(value),
      );

    case MeasureFilterOperation.ChangesBy:
      // Δ<field> < -<value> && Δ<field> > <value>
      // or
      // Δ%<field> < -<value> && Δ%<field> > <value>
      return createOrExpression([
        createBinaryExpression(
          measureFilter.measure + comparisonSuffix,
          V1Operation.OPERATION_LT,
          -Math.abs(value),
        ),
        createBinaryExpression(
          measureFilter.measure + comparisonSuffix,
          V1Operation.OPERATION_GT,
          value,
        ),
      ]);
  }
}
