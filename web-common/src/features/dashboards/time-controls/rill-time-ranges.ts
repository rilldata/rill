import { validateRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { get } from "svelte/store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  getQueryServiceMetricsViewTimeRangesQueryKey,
  queryServiceMetricsViewTimeRanges,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

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

    rillTimeToTimeRange.set(rillTimes.length, i);
    rillTimes.push(tr.name);
  });

  if (rillTimes.length === 0) return timeRangesToReturn;

  const instanceId = get(runtime).instanceId;
  const metricsViewName = exploreSpec.metricsView!;

  try {
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
  } catch (error) {
    console.error(
      `Failed to resolve time ranges for metrics view ${metricsViewName} in instance ${instanceId}`,
      error,
    );
    return timeRangesToReturn;
  }
}
