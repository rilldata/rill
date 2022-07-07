import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  BasicMeasureDefinition,
  getFallbackMeasureName,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export function getFilterFromFilters(filters: ActiveValues): string {
  return Object.keys(filters)
    .map((field) => {
      return filters[field]
        .map(([value, filterType]) =>
          filterType ? `"${field}" = '${value}'` : `"${field}" != '${value}'`
        )
        .join(" OR ");
    })
    .join(" AND ");
}

export function normaliseMeasures(measures: Array<BasicMeasureDefinition>) {
  if (!measures) return [{ expression: "count(*)", id: "", sqlName: "count" }];
  measures.forEach((measure, idx) => {
    if (!measure.sqlName) {
      measure.sqlName = getFallbackMeasureName(idx);
    }
  });
  return measures;
}

export function getExpressionColumnsFromMeasures(
  measures: Array<BasicMeasureDefinition>
): string {
  return measures
    .map((measure) => `${measure.expression} as ${measure.sqlName}`)
    .join(", ");
}

export function getCoalesceStatementsMeasures(
  measures: Array<BasicMeasureDefinition>
): string {
  return measures
    .map(
      (measure) =>
        `COALESCE(series.${measure.sqlName}, 0) as ${measure.sqlName}`
    )
    .join(", ");
}
