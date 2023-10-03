import { faker } from "@faker-js/faker";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import {
  createQueryServiceMetricsViewAggregation,
  V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import { range } from "./util";

export type TimeDimensionDataState = {
  isFetching: boolean;
  comparing: "dimension" | "time" | "none";
  data?: unknown[];
};

export type TimeSeriesDataStore = Readable<TimeDimensionDataState>;

function createTimeDimensionAggregation(
  ctx: StateManagers
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          dimensions: [{ name: dashboard?.selectedComparisonDimension }],
          measures: [{ name: dashboard?.expandedMeasureName }],
          filter: dashboard?.filters,
          timeStart: timeControls.adjustedStart,
          timeEnd: timeControls.adjustedEnd,
          limit: "1000", //TODO Use blocks later
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set)
  );
}

// Move to Data Graphics
const scale = (n: number) => (n * 13).toFixed(2);
const createSpark = (
  nums: number[]
) => `<svg width="34" height="13" viewBox="0 0 34 13" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M1 ${scale(nums[0])}L5.5 ${scale(nums[1])}L11.5 ${scale(
  nums[2]
)}L17 ${scale(nums[3])}L21 ${scale(nums[4])}L28 ${scale(nums[5])}L33 ${scale(
  nums[6]
)}" stroke="#9CA3AF"/>
</svg>`;

function prepareDimensionData(data) {
  if (!data) return;

  const rowCount = data?.length;
  // When using row headers, be careful not to accidentally merge cells
  const rowHeaderData = range(0, rowCount, (_) => [
    // Dim
    {
      value: data.name,
    },
    // Measure total
    {
      value: 23.4,
      spark: createSpark(range(0, 7, (_) => Math.random())),
    },
    // Measure percent of total
    {
      value: 44 + "%",
    },
  ]);

  return {
    rowCount,
    rowHeaderData,
  };
}

function prepareTimeComparisonDate(data) {
  if (!data) {
    return;
  }
  // todo add
}

export function createTimeDimensionDataStore(ctx: StateManagers) {
  return derived(
    [ctx.dashboardStore, useTimeControlStore(ctx), useTimeSeriesDataStore(ctx)],
    ([dashboardStore, timeControls, timeSeries]) => {
      let comparing;
      let data;
      if (dashboardStore?.selectedComparisonDimension) {
        comparing = "dimension";
        data = prepareDimensionData(timeSeries?.dimensionTableData);
      } else if (timeControls.showComparison) {
        comparing = "time";
        data = prepareTimeComparisonDate(timeSeries?.timeSeriesData);
      } else {
        comparing = "none";
      }

      return { comparing, data };
    }
  ) as TimeSeriesDataStore;
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeDimensionDataStore =
  memoizeMetricsStore<TimeSeriesDataStore>((ctx: StateManagers) =>
    createTimeDimensionDataStore(ctx)
  );
