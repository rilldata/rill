import {
  BasicMeasureDefinition,
  getFallbackMeasureName,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { MetricsViewRequestFilter } from "$common/rill-developer-service/MetricsViewActions";
import type { ActiveValues } from "$lib/application-state-stores/explorer-stores";

export function getFilterFromFilters(filters: ActiveValues): string {
  return Object.keys(filters)
    .map((field) => {
      return (
        "(" +
        filters[field]
          .map(([value, filterType]) => {
            if (value == null) {
              return filterType
                ? `"${field}" IS NULL`
                : `"${field}" IS NOT NULL`;
            } else {
              return filterType
                ? `"${field}" = '${value}'`
                : `"${field}" != '${value}'`;
            }
          })
          .join(" OR ") +
        ")"
      );
    })
    .join(" AND ");
}

export function getFilterFromMetricsViewFilters(
  filters: MetricsViewRequestFilter
): string {
  const includeFilters = filters.include
    .map((dimensionValues) =>
      dimensionValues.values
        .map((value) =>
          value === null
            ? `"${dimensionValues.name}" IS NULL`
            : `"${dimensionValues.name}" = '${value}'`
        )
        .join(" OR ")
    )
    .map((filter) => `(${filter})`)
    .join(" AND ");

  const excludeFilters = filters.exclude
    .map((dimensionValues) =>
      dimensionValues.values
        .map((value) =>
          value === null
            ? `"${dimensionValues.name}" IS NOT NULL`
            : `"${dimensionValues.name}" != '${value}'`
        )
        .join(" OR ")
    )
    .map((filter) => `(${filter})`)
    .join(" AND ");
  return [includeFilters, excludeFilters].filter(Boolean).join(" AND ");
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

export function getCoalesceExpressionForMeasures(
  measures: Array<BasicMeasureDefinition>
): string {
  return measures
    .map(
      (measure) => `COALESCE(${measure.expression}, 0) as ${measure.sqlName}`
    )
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

// TODO: remove ActiveValues once all uses have been moved
export function getWhereClauseFromFilters(
  metricViewFilters: MetricsViewRequestFilter,
  timestampColumn: string,
  timeRange: TimeSeriesTimeRange,
  prefix: string
) {
  const whereClauses = [];
  if (
    metricViewFilters?.include?.length ||
    metricViewFilters?.exclude?.length
  ) {
    whereClauses.push(getFilterFromMetricsViewFilters(metricViewFilters));
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
  timeRange = normaliseTimeRange(timeRange);
  if (timeRange.start) {
    timeRangeFilters.push(
      `"${timestampColumn}" >= TIMESTAMP '${timeRange.start}'`
    );
  }
  if (timeRange.end) {
    timeRangeFilters.push(
      `"${timestampColumn}" <= TIMESTAMP '${timeRange.end}'`
    );
  }
  return timeRangeFilters.join(" AND ");
}

function normaliseTimeRange(timeRange: TimeSeriesTimeRange) {
  const returnTimeRange: TimeSeriesTimeRange = {
    ...(timeRange.interval ? { interval: timeRange.interval } : {}),
  };
  if (timeRange.start) {
    const startDate = new Date(timeRange.start);
    if (!Number.isNaN(startDate.getTime())) {
      returnTimeRange.start = startDate.toISOString();
    }
  }
  if (timeRange.end) {
    const endDate = new Date(timeRange.end);
    if (!Number.isNaN(endDate.getTime())) {
      returnTimeRange.end = endDate.toISOString();
    }
  }
  return returnTimeRange;
}
