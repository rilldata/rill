import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  BasicMeasureDefinition,
  getFallbackMeasureName,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export function getFilterFromFilters(filters: ActiveValues): string {
  return Object.keys(filters)
    .map((field) => {
      return (
        "(" +
        filters[field]
          .map(([value, filterType]) =>
            filterType ? `"${field}" = '${value}'` : `"${field}" != '${value}'`
          )
          .join(" OR ") +
        ")"
      );
    })
    .join(" AND ");
}

/** Sets the sqlName to a fallback measure name, if sqlName is not defined */
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

export function getWhereClauseFromFilters(
  filters: ActiveValues,
  timestampColumn: string,
  timeRange: TimeSeriesTimeRange,
  prefix: string
) {
  const whereClauses = [];
  if (filters && Object.keys(filters).length) {
    whereClauses.push(getFilterFromFilters(filters));
  }
  if (timeRange?.start || timeRange?.end) {
    whereClauses.push(getFilterFromTimeRange(timestampColumn, timeRange));
  }
  return whereClauses.length ? `${prefix} ${whereClauses.join(" AND ")}` : "";
}

export function getFilterFromTimeRange(
  timestampColumn: string,
  timeRange: TimeSeriesTimeRange
): string {
  const timeRangeFilters = new Array<string>();
  if (timeRange.start) {
    timeRangeFilters.push(
      `${timestampColumn} >= TIMESTAMP '${timeRange.start}'`
    );
  }
  if (timeRange.end) {
    timeRangeFilters.push(`${timestampColumn} <= TIMESTAMP '${timeRange.end}'`);
  }
  return timeRangeFilters.join(" AND ");
}
