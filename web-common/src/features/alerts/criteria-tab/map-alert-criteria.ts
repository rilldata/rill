import type { AlertCriteria } from "@rilldata/web-common/features/alerts/form-utils";
import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterComparisonType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { type V1Expression } from "@rilldata/web-common/runtime-client";

// TODO: once we have a design for comparison measure filter in dashboard we should merge the definiton and get rid of this mapping

export function mapAlertCriteriaToExpression(
  criteria: AlertCriteria,
): V1Expression | undefined {
  return mapMeasureFilterToExpr({
    measure: criteria.field,
    comparison: MeasureFilterComparisonType.None,
    operation: criteria.operation,
    value1: criteria.value,
    value2: "",
    not: criteria.not,
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
      value: "0",
      not: false,
    };
  }

  return {
    field: measureFilter?.measure ?? "",
    operation: measureFilter.operation,
    value: measureFilter.value1,
    not: false,
  };
}
