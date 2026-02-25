import { validateRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import {
  getQueryServiceMetricsViewTimeRangesQueryKey,
  queryServiceMetricsViewTimeRanges,
} from "@rilldata/web-common/runtime-client/v2/gen";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function resolveTimeRanges(
  client: RuntimeClient,
  exploreSpec: V1ExploreSpec,
  timeRanges: (DashboardTimeControls | undefined)[],
  timeZone: string | undefined,
  executionTime: string | undefined = undefined,
  timeDimension: string | undefined = undefined,
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

  const metricsViewName = exploreSpec.metricsView!;

  try {
    const timeRangesResp = await fetchTimeRanges({
      client,
      metricsViewName,
      rillTimes,
      timeZone,
      timeDimension,
      executionTime,
    });

    timeRangesResp.resolvedTimeRanges?.forEach((tr, index) => {
      const mappedIndex = rillTimeToTimeRange.get(index);
      if (mappedIndex === undefined || !timeRangesToReturn[mappedIndex]) return;
      timeRangesToReturn[mappedIndex].start = new Date(tr.start!);
      timeRangesToReturn[mappedIndex].end = new Date(tr.end!);
    });

    return timeRangesToReturn;
  } catch (error) {
    console.error(
      `Failed to resolve time ranges for metrics view ${metricsViewName} in instance ${client.instanceId}`,
      error,
    );
    return timeRangesToReturn;
  }
}

export async function fetchTimeRanges({
  client,
  metricsViewName,
  rillTimes,
  timeZone,
  timeDimension,
  executionTime,
  cacheBust = false,
}: {
  client: RuntimeClient;
  metricsViewName: string;
  rillTimes: string[];
  timeDimension?: string | undefined;
  timeZone: string | undefined;
  executionTime?: string;
  cacheBust?: boolean;
}) {
  // executionTime is an ISO string; the v2 client uses fromJson() which
  // handles Timestamp fields as RFC 3339 strings at runtime.
  const request = {
    metricsViewName,
    expressions: rillTimes,
    timeZone,
    executionTime,
    priority: 100,
    timeDimension,
  } as Parameters<typeof queryServiceMetricsViewTimeRanges>[1];

  const queryKey = getQueryServiceMetricsViewTimeRangesQueryKey(
    client.instanceId,
    request,
  );

  if (cacheBust) {
    await queryClient.invalidateQueries({
      queryKey: queryKey,
    });
  }

  const response = await queryClient.fetchQuery({
    queryKey: queryKey,
    queryFn: () => queryServiceMetricsViewTimeRanges(client, request),
    staleTime: 60,
  });

  return response;
}
