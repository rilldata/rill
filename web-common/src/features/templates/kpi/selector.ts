import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeRange,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export function useKPITotals(
  instanceId: string,
  metricViewName: string,
  measure: string,
  timeRange: string,
) {
  return createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [{ name: measure }],
      timeRange: { isoDuration: timeRange },
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

export function useStartEndTime(
  instanceId: string,
  metricViewName: string,
  timeRange: string,
) {
  return createQueryServiceMetricsViewTimeRange(
    instanceId,
    metricViewName,
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
  metricViewName: string,
  measure: string,
  timeRange: string,
  queryClient,
) {
  const allTimeRangeQuery = useMetricsViewTimeRange(instanceId, metricViewName);

  return derived(allTimeRangeQuery, (allTimeRange, set) => {
    if (!allTimeRange.data?.timeRangeSummary?.max) {
      return undefined;
    }

    const maxTime = new Date(allTimeRange.data.timeRangeSummary.max);
    const { startTime, endTime } = isoDurationToTimeRange(timeRange, maxTime);
    const defaultGrain = getDefaultTimeGrain(startTime, endTime);
    return createQueryServiceMetricsViewTimeSeries(
      instanceId,
      metricViewName,
      {
        measureNames: [measure],
        timeStart: startTime.toISOString(),
        timeEnd: endTime.toISOString(),
        timeGranularity: defaultGrain,
      },
      {
        query: {
          select: (data) =>
            data.data?.map((d) => {
              return {
                ts: new Date(d.ts),
                [measure]: d?.records?.[measure],
              };
            }) ?? [],
          queryClient,
        },
      },
    ).subscribe(set);
  });
}
