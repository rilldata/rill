import { createTimeAndFilterStore } from "@rilldata/web-common/features/canvas/components/time-filter-store";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeSeries,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { type CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function useKPITotals(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  measure: string,
  overrideTimeRange: string | undefined,
): CreateQueryResult<number | null, HTTPError> {
  const { selectedTimeRange } = ctx.canvasEntity.timeControls;

  const timeAndFilterStore = createTimeAndFilterStore(
    ctx,
    instanceId,
    metricsViewName,
    {
      timeRangeStore: selectedTimeRange,
      overrideTimeRange: overrideTimeRange,
    },
  );

  return derived(timeAndFilterStore, ({ timeRange, where }, set) => {
    return createQueryServiceMetricsViewAggregation(
      instanceId,
      metricsViewName,
      {
        measures: [{ name: measure }],
        timeRange,
        where,
      },
      {
        query: {
          enabled:
            !!overrideTimeRange || (!!timeRange?.start && !!timeRange?.end),
          select: (data) => {
            return data.data?.[0]?.[measure] ?? null;
          },
          queryClient: ctx.queryClient,
        },
      },
    ).subscribe(set);
  });
}

export function useKPIComparisonTotal(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  measure: string,
  overrideComparisonRange: string | undefined,
): CreateQueryResult<number | null, HTTPError> {
  const { showTimeComparison, selectedComparisonTimeRange } =
    ctx.canvasEntity.timeControls;

  // Build the store that yields { finalTimeRange, where }
  const timeAndFilterStore = createTimeAndFilterStore(
    ctx,
    instanceId,
    metricsViewName,
    {
      timeRangeStore: selectedComparisonTimeRange,
      overrideTimeRange: overrideComparisonRange,
    },
  );

  return derived(
    [timeAndFilterStore, showTimeComparison],
    ([{ timeRange, where }, showComparison], set) => {
      // TODO: Use all time range and then calculate the comparison range

      return createQueryServiceMetricsViewAggregation(
        instanceId,
        metricsViewName,
        {
          measures: [{ name: measure }],
          timeRange,
          where,
        },
        {
          query: {
            enabled:
              !!overrideComparisonRange ||
              (showComparison && !!timeRange?.start && !!timeRange?.end),
            select: (data) => {
              return data.data?.[0]?.[measure] ?? null;
            },
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}

export function useKPISparkline(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  measure: string,
  overrideTimeRange: string | undefined,
): CreateQueryResult<Array<Record<string, unknown>>> {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    { query: { queryClient: ctx.queryClient } },
  );
  const { selectedTimeRange } = ctx.canvasEntity.timeControls;

  const timeAndFilterStore = createTimeAndFilterStore(
    ctx,
    instanceId,
    metricsViewName,
    {
      timeRangeStore: selectedTimeRange,
      overrideTimeRange: overrideTimeRange,
    },
  );

  return derived(
    [allTimeRangeQuery, selectedTimeRange, timeAndFilterStore],
    ([allTimeRange, selectedRange, { timeRange, where }], set) => {
      const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
      const maxTimeDate = new Date(maxTime ?? 0);

      let { start, end } = timeRange;

      let defaultGrain = selectedRange?.interval || V1TimeGrain.TIME_GRAIN_DAY;

      if (overrideTimeRange) {
        const overrideRange = isoDurationToTimeRange(
          overrideTimeRange,
          maxTimeDate,
        );

        defaultGrain = getDefaultTimeGrain(
          overrideRange.startTime,
          overrideRange.endTime,
        );
        start = overrideRange.startTime.toISOString();
        end = overrideRange.endTime.toISOString();
      }

      return createQueryServiceMetricsViewTimeSeries(
        instanceId,
        metricsViewName,
        {
          measureNames: [measure],
          timeStart: start,
          timeEnd: end,
          timeGranularity: defaultGrain,
          timeZone: timeRange.timeZone,
          where,
        },
        {
          query: {
            enabled: !!start && !!end && !!maxTime,
            select: (data) =>
              data.data?.map((d) => ({
                ts: new Date(d.ts as string),
                [measure]: d?.records?.[measure],
              })) ?? [],
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}
