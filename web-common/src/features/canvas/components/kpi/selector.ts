import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeRange,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import { type CreateQueryResult, QueryClient } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function useKPITotals(
  instanceId: string,
  metricsViewName: string,
  measure: string,
  timeRange: string,
  whereSql: string | undefined,
) {
  return createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measure }],
      timeRange: { isoDuration: timeRange },
      whereSql,
    },
    {
      query: {
        select: (data) => {
          return data.data?.[0]?.[measure] ?? null;
        },
      },
    },
  );
}

export function useKPIComparisonTotal(
  instanceId: string,
  metricsViewName: string,
  measure: string,
  comparisonRange: string | undefined,
  timeRange: string,
  whereSql: string | undefined,
  queryClient: QueryClient,
): CreateQueryResult<number | undefined> {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    { query: { queryClient } },
  );

  return derived(allTimeRangeQuery, (allTimeRange, set) => {
    const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
    const maxTimeDate = new Date(maxTime ?? 0);
    const { startTime } = isoDurationToTimeRange(timeRange, maxTimeDate);

    let comparisonStartTime: Date, comparisonEndTime: Date;

    if (comparisonRange) {
      ({ startTime: comparisonStartTime, endTime: comparisonEndTime } =
        isoDurationToTimeRange(comparisonRange, startTime));
    } else {
      comparisonStartTime = new Date(0);
      comparisonEndTime = startTime;
    }

    return createQueryServiceMetricsViewAggregation(
      instanceId,
      metricsViewName,
      {
        measures: [{ name: measure }],
        timeRange: {
          start: comparisonStartTime.toISOString(),
          end: comparisonEndTime.toISOString(),
        },
        whereSql,
      },
      {
        query: {
          queryClient,
          select: (data) => {
            return data.data?.[0]?.[measure] ?? undefined;
          },
          enabled: !!comparisonRange,
        },
      },
    ).subscribe(set);
  });
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
  instanceId: string,
  metricsViewName: string,
  measure: string,
  timeRange: string,
  whereSql: string | undefined,
  queryClient: QueryClient,
): CreateQueryResult<Array<Record<string, unknown>>> {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    { query: { queryClient } },
  );

  return derived(allTimeRangeQuery, (allTimeRange, set) => {
    const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
    const maxTimeDate = new Date(maxTime ?? 0);
    const { startTime, endTime } = isoDurationToTimeRange(
      timeRange,
      maxTimeDate,
    );
    const defaultGrain = getDefaultTimeGrain(startTime, endTime);
    return createQueryServiceMetricsViewTimeSeries(
      instanceId,
      metricsViewName,
      {
        measureNames: [measure],
        timeStart: startTime.toISOString(),
        timeEnd: endTime.toISOString(),
        timeGranularity: defaultGrain,
        whereSql,
      },
      {
        query: {
          enabled: !!startTime && !!endTime && !!maxTime,
          select: (data) =>
            data.data?.map((d) => {
              return {
                ts: new Date(d.ts as string),
                [measure]: d?.records?.[measure],
              };
            }) ?? [],
          queryClient,
        },
      },
    ).subscribe(set);
  });
}
