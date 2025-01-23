import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
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
  componentTimeRange: string | undefined,
  componentFilter: string | undefined,
): CreateQueryResult<number | null, HTTPError> {
  const { canvasEntity } = ctx;

  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentTimeRange,
      componentFilter,
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
            !!componentTimeRange || (!!timeRange?.start && !!timeRange?.end),
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
  componentFilter: string | undefined,
): CreateQueryResult<number | null, HTTPError> {
  const { canvasEntity } = ctx;
  const { showTimeComparison, selectedComparisonTimeRange } =
    canvasEntity.timeControls;

  // Build the store that yields { finalTimeRange, where }
  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      timeRangeStore: selectedComparisonTimeRange,
      componentTimeRange: overrideComparisonRange,
      componentFilter,
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
  componentTimeRange: string | undefined,
  componentFilter: string | undefined,
): CreateQueryResult<Array<Record<string, unknown>>> {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    { query: { queryClient: ctx.queryClient } },
  );
  const { canvasEntity } = ctx;
  const { selectedTimeRange } = canvasEntity.timeControls;

  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentTimeRange: componentTimeRange,
      componentFilter,
    },
  );

  return derived(
    [allTimeRangeQuery, selectedTimeRange, timeAndFilterStore],
    ([allTimeRange, selectedRange, { timeRange, where }], set) => {
      const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
      const maxTimeDate = new Date(maxTime ?? 0);

      let { start, end } = timeRange;
      const { timeZone } = timeRange;

      let defaultGrain = selectedRange?.interval || V1TimeGrain.TIME_GRAIN_DAY;

      if (componentTimeRange) {
        const overrideRange = isoDurationToTimeRange(
          componentTimeRange,
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
          timeZone,
          where,
        },
        {
          query: {
            enabled: !!start && !!end && !!maxTime,
            select: (data) => {
              return prepareTimeSeries(
                data.data || [],
                [],
                TIME_GRAIN[defaultGrain]?.duration,
                timeZone ?? "UTC",
              );
            },
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}
