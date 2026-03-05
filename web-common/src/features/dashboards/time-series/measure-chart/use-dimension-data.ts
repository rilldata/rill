import {
  createAndExpression,
  createInExpression,
  filterExpressions,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationResponse,
  type V1TimeGrain,
  type V1TimeSeriesValue,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  keepPreviousData,
  type CreateQueryResult,
} from "@tanstack/svelte-query";
import { transformAggregateDimensionData, prepareTimeSeries } from "../utils";
import { COMPARISON_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { DateTime } from "luxon";
import type { DimensionSeriesData, TimeSeriesPoint } from "./types";
import type { V1MetricsViewTimeSeriesResponse } from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

/**
 * Creates an aggregation query for dimension comparison data.
 * Used by MeasureChart to fetch per-dimension-value time series.
 */
export function createDimensionAggregationQuery(
  instanceId: string,
  metricsViewName: string,
  measureName: string,
  dimensionName: string,
  dimensionValues: (string | null)[],
  where: V1Expression | undefined,
  timeDimension: string,
  timeStart: string | undefined,
  timeEnd: string | undefined,
  timeGranularity: V1TimeGrain,
  timeZone: string,
  enabled: boolean,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const baseFilter = where
    ? (filterExpressions(where, () => true) ?? createAndExpression([]))
    : createAndExpression([]);
  const updatedFilter = createAndExpression([
    ...(baseFilter.cond?.exprs ?? []),
    createInExpression(dimensionName, dimensionValues),
  ]);

  return createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measureName }],
      dimensions: [
        { name: dimensionName },
        { name: timeDimension, timeGrain: timeGranularity, timeZone },
      ],
      where: sanitiseExpression(updatedFilter, undefined),
      timeRange: {
        start: timeStart,
        end: timeEnd,
        timeDimension,
      },
      sort: [
        { desc: true, name: measureName },
        { desc: false, name: timeDimension },
      ],
      // Upper bound: dimensions × time-grain buckets. Matches the limit
      // used in multiple-dimension-queries.ts. Results exceeding this are
      // silently truncated, which is acceptable since the leaderboard caps
      // visible dimensions well below this threshold.
      limit: "10000",
      offset: "0",
    },
    {
      query: {
        enabled: enabled && dimensionValues.length > 0,
        placeholderData: keepPreviousData,
      },
    },
    queryClient,
  );
}

/**
 * Pure function: transforms aggregation response data into DimensionSeriesData[].
 */
export function buildDimensionSeriesData(
  measureName: string,
  dimensionName: string,
  dimensionValues: (string | null)[],
  timeDimension: string,
  timeGranularity: V1TimeGrain,
  timeZone: string,
  primaryTimeSeriesData: V1MetricsViewTimeSeriesResponse["data"],
  aggData: V1MetricsViewAggregationResponse["data"],
  comparisonTimeSeriesData: V1MetricsViewTimeSeriesResponse["data"] | undefined,
  compAggData: V1MetricsViewAggregationResponse["data"] | undefined,
  isFetching: boolean,
): DimensionSeriesData[] {
  if (!dimensionValues.length || !primaryTimeSeriesData?.length) return [];

  const measures = [measureName];

  const transformedData = transformAggregateDimensionData(
    timeDimension,
    dimensionName,
    measures,
    dimensionValues,
    primaryTimeSeriesData,
    aggData || [],
  );

  let comparisonData: V1TimeSeriesValue[][] = [];
  if (comparisonTimeSeriesData && compAggData) {
    comparisonData = transformAggregateDimensionData(
      timeDimension,
      dimensionName,
      measures,
      dimensionValues,
      comparisonTimeSeriesData,
      compAggData,
    );
  }

  const grainDuration = TIME_GRAIN[timeGranularity]?.duration;
  const results: DimensionSeriesData[] = [];

  for (let i = 0; i < dimensionValues.length; i++) {
    const prepData = prepareTimeSeries(
      transformedData[i],
      comparisonData[i],
      grainDuration,
      timeZone,
    );

    const data: TimeSeriesPoint[] = prepData.map((datum) => {
      const compKey = `comparison.${measureName}`;
      const compTsKey = "comparison.ts";
      return {
        ts: datum.ts
          ? DateTime.fromJSDate(datum.ts, { zone: timeZone })
          : DateTime.invalid("missing"),
        value: (datum[measureName] as number | null) ?? null,
        comparisonValue:
          compKey in datum
            ? ((datum[compKey] as number | null) ?? null)
            : undefined,
        comparisonTs:
          compTsKey in datum && datum[compTsKey]
            ? DateTime.fromJSDate(datum[compTsKey] as Date, {
                zone: timeZone,
              })
            : undefined,
      };
    });

    results.push({
      dimensionValue: dimensionValues[i],
      color: COMPARISON_COLORS[i % COMPARISON_COLORS.length] || "",
      data,
      isFetching,
    });
  }

  return results;
}
