import { CompareWith } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import type { AlertCriteria } from "@rilldata/web-common/features/alerts/form-utils";
import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterComparisonType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  IsCompareMeasureFilterOperation,
  MeasureFilterOperation,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { type V1Expression } from "@rilldata/web-common/runtime-client";

// TODO: once we have a design for comparison measure filter in dashboard we should merge the definiton and get rid of this mapping

export function mapAlertCriteriaToExpression(
  criteria: AlertCriteria,
): V1Expression | undefined {
  let comparison = MeasureFilterComparisonType.None;
  if (criteria.compareWith === CompareWith.Percent) {
    comparison = MeasureFilterComparisonType.PercentageComparison;
  } else if (criteria.operation in IsCompareMeasureFilterOperation) {
    comparison = MeasureFilterComparisonType.AbsoluteComparison;
  }
  return mapMeasureFilterToExpr({
    measure: criteria.field,
    comparison,
    operation: criteria.operation,
    value1: criteria.value,
    value2: "",
  });
}

export function mapExpressionToAlertCriteria(
  expr: V1Expression,
): AlertCriteria {
  const measureFilter = mapExprToMeasureFilter(expr);
  if (!measureFilter) {
    return {
      field: "",
      operation: MeasureFilterOperation.GreaterThan,
      compareWith: CompareWith.Value,
      value: "0",
    };
  }

  return {
    field: measureFilter?.measure ?? "",
    operation: measureFilter.operation,
    compareWith:
      measureFilter.comparison ===
      MeasureFilterComparisonType.PercentageComparison
        ? CompareWith.Percent
        : CompareWith.Value,
    value: measureFilter.value1,
  };
}
