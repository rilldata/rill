import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { includedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import {
  createAndExpression,
  createInExpression,
  filterExpressions,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { createBatches } from "@rilldata/web-common/lib/arrayUtils";
import { type Readable, derived } from "svelte/store";

import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  type TimeSeriesDatum,
  createMetricsViewTimeSeries,
} from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type V1Expression,
  type V1MetricsViewAggregationResponse,
  V1TimeGrain,
  type V1TimeSeriesValue,
  createQueryServiceMetricsViewAggregation,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  type CreateQueryResult,
  keepPreviousData,
} from "@tanstack/svelte-query";
import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
import { dimensionSearchText } from "../stores/dashboard-stores";
import {
  getFilterForComparedDimension,
  prepareTimeSeries,
  transformAggregateDimensionData,
} from "./utils";

const MAX_TDD_VALUES_LENGTH = 250;
const BATCH_SIZE = 50;
export interface DimensionDataItem {
  value: string | null;
  total?: number;
  color: string;
  data: TimeSeriesDatum[];
  isFetching: boolean;
}

interface DimensionTopList {
  values: (string | null)[];
  filter: V1Expression;
  totals?: number[];
}

/***
 * Returns a list of dimension values which for which to fetch
 * timeseries data for a given dimension.
 *
 * For Explore Page -
 * Use the included values if present,
 * otherwise fetch the top values for the dimension
 *
 * For Time Dimension Detail Page -
 * Fetch all the top n values for the dimension
 * and further filter using search text if present
 */

export function getDimensionValuesForComparison(
  ctx: StateManagers,
  measures: string[],
  surface: "chart" | "table",
): Readable<DimensionTopList> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      dimensionSearchText,
    ],
    ([runtime, name, dashboardStore, timeControls, searchText], set) => {
      const isValidMeasureList =
        measures?.length > 0 && measures?.every((m) => m !== undefined);

      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const showTimeDimensionDetail = Boolean(
        dashboardStore?.activePage ===
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      );

      if (!isValidMeasureList || !dimensionName) return;

      // Values to be compared
      let comparisonValues: (string | null)[] = [];
      if (surface === "chart") {
        const dimensionValues = includedDimensionValues({
          dashboard: dashboardStore,
        })(dimensionName);

        if (dimensionValues?.length) {
          // For TDD view max 11 allowed, for Explore max 7 allowed
          comparisonValues = dimensionValues.slice(
            0,
            showTimeDimensionDetail ? 11 : 7,
          ) as (string | null)[];
        }
        return set({
          values: comparisonValues,
          filter: dashboardStore?.whereFilter,
        });
      } else if (surface === "table") {
        let sortBy = showTimeDimensionDetail
          ? dashboardStore.tdd.expandedMeasureName
          : dashboardStore.leaderboardSortByMeasureName;
        if (dashboardStore?.dashboardSortType === SortType.DIMENSION) {
          sortBy = dimensionName;
        }

        return derived(
          createQueryServiceMetricsViewAggregation(
            runtime.instanceId,
            name,
            {
              measures: measures.map((measure) => ({ name: measure })),
              dimensions: [{ name: dimensionName }],
              where: sanitiseExpression(
                mergeDimensionAndMeasureFilters(
                  getDimensionFilterWithSearch(
                    dashboardStore?.whereFilter,
                    searchText,
                    dimensionName,
                  ),
                  dashboardStore.dimensionThresholdFilters,
                ),
                undefined,
              ),
              timeStart: timeControls.timeStart,
              timeEnd: timeControls.timeEnd,
              sort: [
                {
                  desc:
                    dashboardStore.sortDirection === SortDirection.DESCENDING,
                  name: sortBy,
                },
              ],
              limit: MAX_TDD_VALUES_LENGTH.toString(),
              offset: "0",
            },
            {
              query: {
                enabled:
                  timeControls.ready &&
                  !!dashboardStore?.selectedComparisonDimension,
              },
            },
            ctx.queryClient,
          ),
          (topListData) => {
            if (topListData?.isFetching || !dimensionName)
              return {
                values: [],
                filter: dashboardStore?.whereFilter,
              };
            const columnName =
              topListData?.data?.schema?.fields?.[0]?.name || dimensionName;
            const totalValues = topListData?.data?.data?.map(
              (d) => d[measures[0]],
            ) as number[];
            const topListValues = topListData?.data?.data?.map(
              (d) => d[columnName],
            ) as string[];

            return {
              totals: totalValues,
              values: topListValues?.slice(0, MAX_TDD_VALUES_LENGTH),
              filter: getFilterForComparedDimension(
                dimensionName,
                dashboardStore?.whereFilter,
              ),
            };
          },
        ).subscribe(set);
      }
    },
  );
}

function batchAggregationQueries(
  ctx: StateManagers,
  measures: string[],
  dimensionValues: DimensionTopList,
  includeTimeComparisonForDimension: boolean,
) {
  const batches = createBatches(dimensionValues.values, BATCH_SIZE);
  let queries = batches.map((batch) =>
    getAggregationQueryForTopList(ctx, measures, {
      values: batch,
      filter: dimensionValues.filter,
    }),
  );

  if (includeTimeComparisonForDimension) {
    queries = queries.concat(
      batches.map((batch) =>
        getAggregationQueryForTopList(
          ctx,
          measures,
          {
            values: batch,
            filter: dimensionValues.filter,
          },
          true,
        ),
      ),
    );
  }

  return { batchedTopList: batches, batchedQueries: queries };
}

function getAggregationQueryForTopList(
  ctx: StateManagers,
  measures: string[],
  dimensionValues: DimensionTopList,
  isTimeComparison: boolean = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricsViewName, dashboardStore, timeStore], set) => {
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const timeGrain =
        timeStore?.selectedTimeRange?.interval || V1TimeGrain.TIME_GRAIN_DAY;
      const timeZone = dashboardStore?.selectedTimezone;
      const timeDimension = timeStore?.timeDimension;
      const topListValues = dimensionValues?.values || [];

      if (!topListValues.length || !dimensionName) return;

      const updatedFilter =
        filterExpressions(dimensionValues?.filter, () => true) ??
        createAndExpression([]);
      updatedFilter.cond?.exprs?.push(
        createInExpression(dimensionName, topListValues),
      );

      return createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures.map((measure) => ({ name: measure })),
          dimensions: [
            { name: dimensionName },
            { name: timeDimension, timeGrain, timeZone },
          ],
          where: sanitiseExpression(updatedFilter, undefined),
          timeStart: isTimeComparison
            ? timeStore?.comparisonAdjustedStart
            : timeStore?.adjustedStart,
          timeEnd: isTimeComparison
            ? timeStore?.comparisonAdjustedEnd
            : timeStore?.adjustedEnd,
          sort: [
            {
              desc: dashboardStore.sortDirection === SortDirection.DESCENDING,
              name: measures[0],
            },
            { desc: false, name: timeDimension },
          ],
          limit: "10000",
          offset: "0",
        },
        {
          query: {
            enabled: !!timeStore.ready && !!ctx.dashboardStore,
            placeholderData: keepPreviousData,
          },
        },
        ctx.queryClient,
      ).subscribe(set);
    },
  );
}

/***
 * Fetches the timeseries data for a given dimension
 * for a infered set of dimension values and measures
 */
export function getDimensionValueTimeSeries(
  ctx: StateManagers,
  measures: string[],
  surface: "chart" | "table",
): Readable<DimensionDataItem[]> {
  return derived(
    [
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      createMetricsViewTimeSeries(ctx, measures, false),
      createMetricsViewTimeSeries(ctx, measures, true),
      getDimensionValuesForComparison(ctx, measures, surface),
    ],
    (
      [
        dashboardStore,
        timeStore,
        timeSeriesData,
        comparisonTimeSeriesData,
        dimensionValues,
      ],
      set,
    ) => {
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const topListValues = dimensionValues?.values || [];
      const timeGrain =
        timeStore?.selectedTimeRange?.interval || V1TimeGrain.TIME_GRAIN_DAY;
      const timeZone = dashboardStore?.selectedTimezone;
      const timeDimension = timeStore?.timeDimension;
      const isValidMeasureList =
        measures?.length > 0 && measures?.every((m) => m !== undefined);
      const includeTimeComparisonForDimension = Boolean(
        timeStore?.comparisonAdjustedStart && surface === "chart",
      );

      if (
        !topListValues.length ||
        !isValidMeasureList ||
        !dimensionName ||
        timeSeriesData?.isFetching
      )
        return set([]);
      if (!timeDimension || dashboardStore?.selectedScrubRange?.isScrubbing)
        return;

      const { batchedTopList, batchedQueries } = batchAggregationQueries(
        ctx,
        measures,
        dimensionValues,
        includeTimeComparisonForDimension,
      );

      return derived(batchedQueries, (batchedAggTimeSeriesData) => {
        let transformedData: V1TimeSeriesValue[][] = [];
        for (let i = 0; i < batchedTopList.length; i++) {
          transformedData = transformedData.concat(
            transformAggregateDimensionData(
              timeDimension,
              dimensionName,
              measures,
              batchedTopList[i],
              timeSeriesData?.data?.data || [],
              batchedAggTimeSeriesData[i]?.data?.data || [],
            ),
          );
        }

        let comparisonData: V1TimeSeriesValue[][] = [];

        if (includeTimeComparisonForDimension) {
          {
            comparisonData = transformAggregateDimensionData(
              timeDimension,
              dimensionName,
              measures,
              // For chart surface, we only have 1 batch
              batchedTopList[0],
              comparisonTimeSeriesData?.data?.data || [],
              batchedAggTimeSeriesData[1]?.data?.data || [],
            );
          }
        }

        const isFetching = batchedAggTimeSeriesData.some((d) => d.isFetching);

        const results: DimensionDataItem[] = [];
        for (let i = 0; i < topListValues.length; i++) {
          const value = topListValues[i];
          const prepData = prepareTimeSeries(
            transformedData[i],
            comparisonData[i],
            TIME_GRAIN[timeGrain]?.duration,
            timeZone,
          );

          let total;
          if (surface === "table") {
            total = dimensionValues?.totals?.[i];
          }

          results.push({
            value,
            total,
            color: COMPARIONS_COLORS[i] ? COMPARIONS_COLORS[i] : "",
            data: prepData,
            isFetching,
          });
        }

        return results;
      }).subscribe(set);
    },
  );
}
