import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeRange,
  createQueryServiceMetricsViewTimeSeries,
  type V1TimeRange,
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
  const { timeControls } = ctx.canvasEntity;

  return derived(
    [timeControls.selectedTimeRange, timeControls.selectedTimezone],
    ([selectedTimeRange, timeZone], set) => {
      let timeRange: V1TimeRange = {
        start: selectedTimeRange?.start?.toISOString(),
        end: selectedTimeRange?.end?.toISOString(),
        timeZone,
      };

      if (overrideTimeRange) {
        timeRange = { isoDuration: overrideTimeRange, timeZone };
      }
      return createQueryServiceMetricsViewAggregation(
        instanceId,
        metricsViewName,
        {
          measures: [{ name: measure }],
          timeRange,
        },
        {
          query: {
            enabled: !!selectedTimeRange?.start && !!selectedTimeRange?.end,
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

export function useKPIComparisonTotal(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  measure: string,
  overrideComparisonRange: string | undefined,
): CreateQueryResult<number | null, HTTPError> {
  const { timeControls } = ctx.canvasEntity;

  return derived(
    [
      timeControls.selectedComparisonTimeRange,
      timeControls.selectedTimezone,
      timeControls.showTimeComparison,
    ],
    ([selectedComparisonTimeRange, timeZone, showComparison], set) => {
      let timeRange: V1TimeRange = {
        start: selectedComparisonTimeRange?.start?.toISOString(),
        end: selectedComparisonTimeRange?.end?.toISOString(),
        timeZone,
      };

      // TODO: Use all time range and then calculate the comparison range
      if (overrideComparisonRange) {
        timeRange = { isoDuration: overrideComparisonRange, timeZone };
      }
      return createQueryServiceMetricsViewAggregation(
        instanceId,
        metricsViewName,
        {
          measures: [{ name: measure }],
          timeRange,
        },
        {
          query: {
            enabled:
              !!overrideComparisonRange ||
              (showComparison &&
                !!selectedComparisonTimeRange?.start &&
                !!selectedComparisonTimeRange?.end),
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

export function useStartEndTime(
  instanceId: string,
  metricsViewName: string,
  timeRange: string,
) {
  return createQueryServiceMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
    {
      query: {
        select: (data) => {
          const maxTime = new Date(data?.timeRangeSummary?.max ?? 0);
          const { startTime, endTime } = isoDurationToTimeRange(
            timeRange,
            maxTime,
          );

          return { start: startTime, end: endTime };
        },
      },
    },
  );
}

export function useKPISparkline(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  measure: string,
  overrideTimeRange: string | undefined,
  whereSql: string | undefined,
): CreateQueryResult<Array<Record<string, unknown>>> {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    { query: { queryClient: ctx.queryClient } },
  );

  const { timeControls } = ctx.canvasEntity;
  return derived(
    [
      allTimeRangeQuery,
      timeControls.selectedTimeRange,
      timeControls.selectedTimezone,
    ],
    ([allTimeRange, selectedTimeRange, timeZone], set) => {
      const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
      const maxTimeDate = new Date(maxTime ?? 0);
      let startTime = selectedTimeRange?.start;
      let endTime = selectedTimeRange?.end;

      if (overrideTimeRange) {
        const { startTime: start, endTime: end } = isoDurationToTimeRange(
          overrideTimeRange,
          maxTimeDate,
        );
        startTime = start;
        endTime = end;
      }

      const defaultGrain = getDefaultTimeGrain(startTime, endTime);
      return createQueryServiceMetricsViewTimeSeries(
        instanceId,
        metricsViewName,
        {
          measureNames: [measure],
          timeStart: startTime.toISOString(),
          timeEnd: endTime.toISOString(),
          timeGranularity: defaultGrain,
          timeZone,
          whereSql,
        },
        {
          query: {
            enabled: !!startTime && !!endTime && !!maxTime,
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
