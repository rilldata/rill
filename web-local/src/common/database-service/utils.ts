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
      if (nonNullValues.length > 0) {
        conditions.push(
          `"${dimensionValue.name}" ${prefix} IN (${nonNullValues
            .map((value) => `'${escapeFilterValue(value)}'`)
            .join(",")}) `
        );
      }
      if (nonNullValues.length < dimensionValue.in.length) {
        conditions.push(`"${dimensionValue.name}" IS ${prefix} NULL`);
      }
      if (dimensionValue.like?.length) {
        conditions.push(
          ...dimensionValue.like.map(
            (value) =>
              `"${dimensionValue.name}" ${prefix} ILIKE '${escapeFilterValue(
                value
              )}'`
          )
        );
      }
      return conditions.join(` ${dimensionJoiner} `);
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
