import {
  parseRillTime,
  validateRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { GrainAliasToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
import {
  getQueryServiceMetricsViewTimeRangesQueryOptions,
  type V1ExploreSpec,
  type V1MetricsViewTimeRangesResponse,
  type V1TimeGrain,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

// Ranges anchored at `now` bypass this by invalidating before they fetch.
const TimeRangesStaleTimeMs = 60_000;

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
  const queryOptions = getTimeRangesQueryOptions(client, {
    metricsViewName,
    expressions: rillTimes,
    timeZone,
    timeDimension,
    executionTime,
  });

  if (cacheBust) {
    await queryClient.invalidateQueries({ queryKey: queryOptions.queryKey });
  }

  return queryClient.fetchQuery(queryOptions);
}

/**
 * Query options for resolving rill time expressions against a metrics view.
 * Shared by imperative fetches and subscriptions so both hit the same cache entry.
 */
export function getTimeRangesQueryOptions(
  client: RuntimeClient,
  {
    metricsViewName,
    expressions,
    timeZone,
    timeDimension,
    executionTime,
  }: {
    metricsViewName: string;
    expressions: string[];
    timeZone: string | undefined;
    timeDimension?: string | undefined;
    executionTime?: string | undefined;
  },
) {
  return getQueryServiceMetricsViewTimeRangesQueryOptions(
    client,
    {
      metricsViewName,
      expressions,
      timeZone,
      executionTime: executionTime as any,
      timeDimension,
    },
    { query: { staleTime: TimeRangesStaleTimeMs } },
  );
}

export type DerivedInterval = {
  interval: Interval;
  grain?: V1TimeGrain | undefined;
  fullTimeRange?: V1TimeRangeSummary;
  error?: string;
};

/**
 * Maps a time ranges response to the interval resolved for `expression`.
 * Returns an invalid interval when the response has no resolved range for the expression.
 */
export function toDerivedInterval(
  response: V1MetricsViewTimeRangesResponse,
  expression: string,
  timeZone: string,
): DerivedInterval {
  const timeRange = response.resolvedTimeRanges?.find(
    (rtr) => rtr.expression === expression,
  );
  if (!timeRange?.start || !timeRange?.end) {
    return { interval: Interval.invalid("Invalid time range") };
  }

  let grain: V1TimeGrain | undefined;
  try {
    const parsed = parseRillTime(expression);
    grain = parsed.asOfLabel?.snap
      ? GrainAliasToV1TimeGrain[parsed.asOfLabel.snap]
      : parsed.rangeGrain;
  } catch {
    // The runtime resolved the expression, so only the grain hint is lost.
  }

  return {
    interval: Interval.fromDateTimes(
      DateTime.fromISO(timeRange.start).setZone(timeZone),
      DateTime.fromISO(timeRange.end).setZone(timeZone),
    ),
    fullTimeRange: response.fullTimeRange,
    grain,
  };
}
