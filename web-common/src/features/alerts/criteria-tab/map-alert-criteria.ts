import type { AlertCriteria } from "@rilldata/web-common/features/alerts/form-utils";
import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterComparisonType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { type V1Expression } from "@rilldata/web-common/runtime-client";

// These are here to maintain differences between filter pill and alert criteria.
// For now, it has no difference, but soon we will have comparison that is only supported in alert criteria.

export function mapAlertCriteriaToExpression(
  criteria: AlertCriteria,
): V1Expression | undefined {
  return mapMeasureFilterToExpr({
    measure: criteria.field,
    comparison: MeasureFilterComparisonType.None,
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
      value: "0",
    };
  }

  return {
    field: measureFilter?.measure ?? "",
    operation: measureFilter.operation,
    value: measureFilter.value1,
  };
}
