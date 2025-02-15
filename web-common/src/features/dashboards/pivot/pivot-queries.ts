import {
  createAndExpression,
  createInExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import {
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
  createQueryServiceMetricsViewAggregation,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { type Readable, derived, readable } from "svelte/store";
import { mergeFilters } from "./pivot-merge-filters";
import {
  getErrorFromResponses,
  getFilterForMeasuresTotalsAxesQuery,
  getTimeGrainFromDimension,
  isTimeDimension,
  prepareMeasureForComparison,
} from "./pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotAxesData,
  type PivotDashboardContext,
  type PivotDataStoreConfig,
  type PivotQueryError,
} from "./types";

/**
 * Wrapper function for Aggregate Query API
 */
export function createPivotAggregationRowQuery(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  measures: V1MetricsViewAggregationMeasure[],
  dimensions: V1MetricsViewAggregationDimension[],
  whereFilter: V1Expression,
  sort: V1MetricsViewAggregationSort[] = [],
  limit = "100",
  offset = "0",
  timeRange: TimeRangeString | undefined = undefined,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  if (!sort.length) {
    sort = [
      {
        desc: false,
        name: measures[0]?.name || dimensions?.[0]?.name,
      },
    ];
  }

  let hasComparison = false;
  const comparisonTime = config.comparisonTime;
  if (
    measures.some(
      (m) =>
        m.name?.endsWith(COMPARISON_PERCENT) ||
        m.name?.endsWith(COMPARISON_DELTA),
    )
  ) {
    hasComparison = true;
  }

  return derived(
    [runtime, ctx.metricsViewName],
    ([$runtime, metricsViewName], set) =>
      createQueryServiceMetricsViewAggregation(
        $runtime.instanceId,
        metricsViewName,
        {
          measures: prepareMeasureForComparison(measures),
          dimensions,
          where: sanitiseExpression(whereFilter, undefined),
          timeRange: {
            start: timeRange?.start ? timeRange.start : config.time.timeStart,
            end: timeRange?.end ? timeRange.end : config.time.timeEnd,
          },
          comparisonTimeRange:
            hasComparison && comparisonTime
              ? {
                  start: comparisonTime.start,
                  end: comparisonTime.end,
                }
              : undefined,
          sort,
          limit,
          offset,
        },
        {
          query: {
            enabled: ctx.enabled,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}

/***
 * Get a list of axis values for a given list of dimension values and filters
 */
export function getAxisForDimensions(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  dimensions: string[],
  measures: V1MetricsViewAggregationMeasure[],
  whereFilter: V1Expression,
  sortBy: V1MetricsViewAggregationSort[] = [],
  timeRange: TimeRangeString | undefined = undefined,
  limit = "100",
  offset = "0",
): Readable<PivotAxesData | null> {
  if (!dimensions.length) return readable(null);

  const { time } = config;

  let sortProvided = true;
  if (!sortBy.length) {
    sortBy = [
      {
        desc: true,
        name: measures[0]?.name || dimensions?.[0],
      },
    ];
    sortProvided = false;
  }

  const dimensionBody = dimensions.map((d) => {
    if (isTimeDimension(d, time.timeDimension)) {
      return {
        name: time.timeDimension,
        timeGrain: getTimeGrainFromDimension(d),
        timeZone: time.timeZone,
        alias: d,
      };
    } else return { name: d };
  });

  return derived(
    dimensionBody.map((dimension) => {
      let sortByForDimension = sortBy;
      if (
        isTimeDimension(dimension.alias, time.timeDimension) &&
        !sortProvided
      ) {
        sortByForDimension = [
          {
            desc: false,
            name: dimension.alias,
          },
        ];
      }
      return createPivotAggregationRowQuery(
        ctx,
        config,
        measures,
        [dimension],
        whereFilter,
        sortByForDimension,
        limit,
        offset,
        timeRange,
      );
    }),
    (data) => {
      const axesMap: Record<string, string[]> = {};
      const totalsMap: Record<
        string,
        V1MetricsViewAggregationResponseDataItem[]
      > = {};

      // Wait for all data to populate
      if (data.some((d) => d?.isFetching)) return { isFetching: true };

      // Check for errors in any of the queries
      const errors: PivotQueryError[] = getErrorFromResponses(data);
      if (errors.length) {
        return {
          isFetching: false,
          error: errors,
        };
      }

      data.forEach((d, i: number) => {
        const dimensionName = dimensions[i];

        axesMap[dimensionName] = (d?.data?.data || [])?.map(
          (dimValue) => dimValue[dimensionName] as string,
        );
        totalsMap[dimensionName] = d?.data?.data || [];
      });

      if (Object.values(axesMap).some((d) => !d)) return { isFetching: true };
      return {
        isFetching: false,
        data: axesMap,
        totals: totalsMap,
      };
    },
  );
}

export function getAxisQueryForMeasureTotals(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  isMeasureSortAccessor: boolean,
  sortAccessor: string | undefined,
  anchorDimension: string,
  rowDimensionValues: string[],
  timeRange: TimeRangeString,
  otherFilters: V1Expression | undefined = undefined,
) {
  let rowAxesQueryForMeasureTotals: Readable<PivotAxesData | null> =
    readable(null);

  if (rowDimensionValues.length && isMeasureSortAccessor && sortAccessor) {
    const { measureNames } = config;
    const measuresBody = measureNames.map((m) => ({ name: m }));

    const sortedRowFilters = getFilterForMeasuresTotalsAxesQuery(
      config,
      anchorDimension,
      rowDimensionValues,
    );

    let mergedFilter: V1Expression | undefined = sortedRowFilters;

    if (otherFilters) {
      mergedFilter = mergeFilters(otherFilters, sortedRowFilters);
    }

    rowAxesQueryForMeasureTotals = getAxisForDimensions(
      ctx,
      config,
      [anchorDimension],
      measuresBody,
      mergedFilter ?? createAndExpression([]),
      [],
      timeRange,
    );
  }

  return rowAxesQueryForMeasureTotals;
}

export function getTotalsRowQuery(
  ctx: PivotDashboardContext,
  config: PivotDataStoreConfig,
  colDimensionAxes: Record<string, string[]> = {},
) {
  const { colDimensionNames } = config;

  const { time } = config;
  const measureBody = config.measureNames.map((m) => ({ name: m }));
  const dimensionBody = colDimensionNames.map((dimension) => {
    if (isTimeDimension(dimension, time.timeDimension)) {
      return {
        name: time.timeDimension,
        timeGrain: getTimeGrainFromDimension(dimension),
        timeZone: time.timeZone,
        alias: dimension,
      };
    } else return { name: dimension };
  });

  const colFilters = colDimensionNames
    .filter((d) => !isTimeDimension(d, time.timeDimension))
    .map((dimension) =>
      createInExpression(dimension, colDimensionAxes[dimension]),
    );

  const mergedFilter =
    mergeFilters(createAndExpression(colFilters), config.whereFilter) ??
    createAndExpression([]);

  const sortBy = [
    {
      desc: true,
      name: config.measureNames[0],
    },
  ];
  return createPivotAggregationRowQuery(
    ctx,
    config,
    measureBody,
    dimensionBody,
    mergedFilter,
    sortBy,
    "300",
  );
}
