import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import {
  normaliseRillTime,
  parseRillTime,
  validateRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { derived, get } from "svelte/store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  createQueryServiceMetricsViewTimeRanges,
  getQueryServiceMetricsViewTimeRangesQueryKey,
  queryServiceMetricsViewTimeRanges,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1MetricsViewTimeRangesResponse,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function fetchTimeRanges(
  exploreSpec: V1ExploreSpec,
  defaultPreset: V1ExplorePreset,
) {
  const rillTimes = dedupe(
    [
      ...(defaultPreset.timeRange ? [defaultPreset.timeRange] : []),
      ...(exploreSpec.timeRanges?.length
        ? exploreSpec.timeRanges.map((t) => t.range!)
        : []),
    ].map(normaliseRillTime),
  );
  const rillTimesWithTimezones = rillTimes.map((tr) => {
    try {
      const rillTime = parseRillTime(tr);
      if (defaultPreset.timezone) {
        rillTime.addTimezone(defaultPreset.timezone);
      }
      return rillTime.toString();
    } catch {
      return tr;
    }
  });

  const instanceId = get(runtime).instanceId;
  const metricsViewName = exploreSpec.metricsView!;

  const timeRangesResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewTimeRangesQueryKey(
      instanceId,
      metricsViewName,
      { expressions: rillTimesWithTimezones },
    ),
    queryFn: () =>
      queryServiceMetricsViewTimeRanges(instanceId, metricsViewName, {
        expressions: rillTimesWithTimezones,
      }),
  });
  return (
    timeRangesResp.timeRanges?.map((tr, i) => ({
      start: tr.start,
      end: tr.end,
      expression: rillTimes[i],
    })) ?? []
  );
}

export function useTimeRanges(instanceId: string, exploreName: string) {
  return derived(
    useExploreValidSpec(instanceId, exploreName),
    (validSpecResp, set) => {
      const metricsViewSpec = validSpecResp.data?.metricsView;
      const exploreSpec = validSpecResp.data?.explore;

      if (!metricsViewSpec || !metricsViewSpec.timeDimension || !exploreSpec) {
        set({
          data: undefined,
          error: null,
          isLoading: false,
          isError: false,
        } as any);
        return;
      }

      const rillTimes = dedupe(
        [
          ...(exploreSpec.defaultPreset?.timeRange
            ? [exploreSpec.defaultPreset.timeRange]
            : []),
          ...(exploreSpec.timeRanges?.length
            ? exploreSpec.timeRanges.map((t) => t.range!)
            : []),
        ].map(normaliseRillTime),
      );

      createQueryServiceMetricsViewTimeRanges(
        instanceId,
        exploreSpec.metricsView!,
        {
          expressions: rillTimes,
          timeZone: exploreSpec.defaultPreset?.timezone,
        },
        undefined,
        queryClient,
      ).subscribe(set);
    },
  ) as CompoundQueryResult<V1MetricsViewTimeRangesResponse>;
}

export async function resolveTimeRanges(
  exploreSpec: V1ExploreSpec,
  timeRanges: (DashboardTimeControls | undefined)[],
  timeZone: string | undefined,
) {
  const rillTimes: string[] = [];
  const rillTimeToTimeRange = new Map<number, number>();
  const timeRangesToReturn = new Array<DashboardTimeControls | undefined>(
    timeRanges.length,
  );

  timeRanges.forEach((tr, i) => {
    timeRangesToReturn[i] = tr;
    if (
      !tr?.name ||
      // already resolved
      tr.start ||
      tr.end ||
      !!validateRillTime(tr.name)
    )
      return;

    const rillTime = parseRillTime(tr.name);
    rillTimeToTimeRange.set(rillTimes.length, i);
    rillTimes.push(rillTime.toString());
  });

  if (rillTimes.length === 0) return timeRangesToReturn;

  const instanceId = get(runtime).instanceId;
  const metricsViewName = exploreSpec.metricsView!;

  const timeRangesResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewTimeRangesQueryKey(
      instanceId,
      metricsViewName,
      { expressions: rillTimes, timeZone },
    ),
    queryFn: () =>
      queryServiceMetricsViewTimeRanges(instanceId, metricsViewName, {
        expressions: rillTimes,
        timeZone,
      }),
    staleTime: Infinity,
  });
  timeRangesResp.timeRanges?.forEach((tr, index) => {
    const mappedIndex = rillTimeToTimeRange.get(index);
    if (mappedIndex === undefined || !timeRangesToReturn[mappedIndex]) return;
    timeRangesToReturn[mappedIndex].start = new Date(tr.start!);
    timeRangesToReturn[mappedIndex].end = new Date(tr.end!);
  });

  return timeRangesToReturn;
}
