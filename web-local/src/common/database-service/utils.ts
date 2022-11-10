import { escapeColumn } from "@rilldata/web-local/common/database-service/columnUtils";
import {
  BasicMeasureDefinition,
  getFallbackMeasureName,
} from "../data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "./DatabaseTimeSeriesActions";
import type {
  MetricsViewDimensionValues,
  MetricsViewRequestFilter,
} from "../rill-developer-service/MetricsViewActions";

function escapeFilterValue(value: unknown) {
  if (typeof value !== "string") return value;
  return value.replace(/'/g, "''");
}

function getFilterFromDimensionValuesFilter(
  dimensionValues: MetricsViewDimensionValues,
  prefix: "" | "NOT",
  dimensionJoiner: "AND" | "OR"
) {
  return dimensionValues
    .map((dimensionValue) => {
      const nonNullValues = dimensionValue.in.filter((value) => value !== null);
      const conditions = [];
      const escapedDimensionName = escapeColumn(dimensionValue.name);
      if (nonNullValues.length > 0) {
        conditions.push(
          `${escapedDimensionName} ${prefix} IN (${nonNullValues
            .map((value) => `'${escapeFilterValue(value)}'`)
            .join(",")}) `
        );
      }
      if (nonNullValues.length < dimensionValue.in.length) {
        conditions.push(`${escapedDimensionName} IS ${prefix} NULL`);
      }
      if (dimensionValue.like?.length) {
        conditions.push(
          ...dimensionValue.like.map(
            (value) =>
              `${escapedDimensionName} ${prefix} ILIKE '${escapeFilterValue(
                value
              )}'`
          )
        );
      }
      return conditions.length > 0
        ? `(${conditions.join(` ${dimensionJoiner} `)})`
        : "";
    })
    .join(" AND ");
}

export function getFilterFromMetricsViewFilters(
  filters: MetricsViewRequestFilter
): string {
  const includeFilters = getFilterFromDimensionValuesFilter(
    filters.include,
    "",
    "OR"
  );

  const excludeFilters = getFilterFromDimensionValuesFilter(
    filters.exclude,
    "NOT",
    "AND"
  );
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
    const filter = getFilterFromMetricsViewFilters(metricViewFilters);
    if (filter !== "") {
      whereClauses.push(filter);
    }
  }
  if (timeRange?.start || timeRange?.end) {
    const tsFilter = getFilterFromTimeRange(timestampColumn, timeRange);
    if (tsFilter !== "") {
      whereClauses.push(tsFilter);
    }
  }
  return whereClauses.length ? `${prefix} ${whereClauses.join(" AND ")}` : "";
}

export function getFilterFromTimeRange(
  timestampColumn: string,
  timeRange: TimeSeriesTimeRange
): string {
  const timeRangeFilters = new Array<string>();
  timeRange = normaliseTimeRange(timeRange);
  const escapedTimestampColumn = escapeColumn(timestampColumn);
  if (timeRange.start) {
    timeRangeFilters.push(
      `${escapedTimestampColumn} >= TIMESTAMP '${timeRange.start}'`
    );
  }
  if (timeRange.end) {
    timeRangeFilters.push(
      `${escapedTimestampColumn} <= TIMESTAMP '${timeRange.end}'`
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
