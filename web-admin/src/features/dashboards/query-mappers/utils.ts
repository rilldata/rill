import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export-utils";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  getQueryServiceMetricsViewToplistQueryKey,
  queryServiceMetricsViewToplist,
  type QueryServiceMetricsViewToplistBody,
  type V1Expression,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

// We are manually sending in duration, offset and round to grain for previous complete ranges.
// This is to map back that split
const PreviousCompleteRangeReverseMap: Record<string, TimeRangePreset> = {};
for (const preset in PreviousCompleteRangeMap) {
  const range: V1TimeRange = PreviousCompleteRangeMap[preset];
  PreviousCompleteRangeReverseMap[
    `${range.isoDuration}_${range.isoOffset}_${range.roundToGrain}`
  ] = preset as TimeRangePreset;
}

export function getSelectedTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  duration: string,
  executionTime: string,
): DashboardTimeControls | undefined {
  let selectedTimeRange: DashboardTimeControls;

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  if (fullRangeKey in PreviousCompleteRangeReverseMap) {
    duration = PreviousCompleteRangeReverseMap[fullRangeKey];
  }

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (duration) {
    selectedTimeRange = isoDurationToFullTimeRange(
      duration,
      new Date(timeRangeSummary.min),
      new Date(executionTime),
    );
  } else {
    return undefined;
  }

  selectedTimeRange.interval = timeRange.roundToGrain;

  return selectedTimeRange;
}

export async function convertExprToToplist(
  queryClient: QueryClient,
  instanceId: string,
  metricsView: string,
  dimensionName: string,
  measureName: string,
  timeRange: DashboardTimeControls | undefined,
  where: V1Expression | undefined,
  exprs: V1Expression[],
) {
  const toplistBody: QueryServiceMetricsViewToplistBody = {
    dimensionName,
    measureNames: [measureName],
    where,
    having: createAndExpression(exprs),
    limit: "250",
    timeStart: timeRange?.start.toISOString(),
    timeEnd: timeRange?.end.toISOString(),
  };
  const toplist = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewToplistQueryKey(
      instanceId,
      metricsView,
      toplistBody,
    ),
    queryFn: () =>
      queryServiceMetricsViewToplist(instanceId, metricsView, toplistBody),
  });
  if (!toplist.data?.length) {
    return undefined;
  }
  return createInExpression(
    dimensionName,
    toplist.data.map((t) => t[dimensionName]),
  );
}
