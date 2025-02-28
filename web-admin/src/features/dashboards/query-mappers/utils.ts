import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  getQueryServiceMetricsViewAggregationQueryKey,
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  queryServiceMetricsViewAggregation,
  type QueryServiceMetricsViewAggregationBody,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewTimeRangeResponse,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

// We are manually sending in duration, offset and round to grain for previous complete ranges.
// This is to map back that split
const PreviousCompleteRangeReverseMap: Record<string, TimeRangePreset> = {};
for (const preset in PreviousCompleteRangeMap) {
  const range: V1TimeRange = PreviousCompleteRangeMap[preset];
  PreviousCompleteRangeReverseMap[
    `${range.isoDuration}_${range.isoOffset}_${range.roundToGrain}`
  ] = preset as TimeRangePreset;
}

export function fillTimeRange(
  dashboard: MetricsExplorerEntity,
  reqTimeRange: V1TimeRange | undefined,
  reqComparisonTimeRange: V1TimeRange | undefined,
  timeRangeSummary: V1TimeRangeSummary,
  executionTime: string,
) {
  if (reqTimeRange) {
    dashboard.selectedTimeRange = getSelectedTimeRange(
      reqTimeRange,
      timeRangeSummary,
      reqTimeRange.isoDuration ?? "",
      executionTime,
    );
  }

  if (reqComparisonTimeRange) {
    if (
      (!reqComparisonTimeRange.isoOffset &&
        reqComparisonTimeRange.isoDuration) ||
      (reqComparisonTimeRange.isoOffset &&
        reqComparisonTimeRange.isoOffset === reqComparisonTimeRange.isoDuration)
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeComparisonOption.CONTIGUOUS,
        start: undefined as unknown as Date,
        end: undefined as unknown as Date,
      };
    } else {
      dashboard.selectedComparisonTimeRange = getSelectedTimeRange(
        reqComparisonTimeRange,
        timeRangeSummary,
        reqComparisonTimeRange.isoOffset,
        executionTime,
      );
      // temporary fix to not lead to an uncaught error.
      // TODO: we should a single custom label when we move to rill-time syntax
      if (
        dashboard.selectedComparisonTimeRange?.name === TimeRangePreset.CUSTOM
      ) {
        dashboard.selectedComparisonTimeRange.name =
          TimeComparisonOption.CUSTOM;
      }
    }

    if (dashboard.selectedComparisonTimeRange) {
      dashboard.selectedComparisonTimeRange.interval =
        dashboard.selectedTimeRange?.interval;
    }
    dashboard.showTimeComparison = true;
  }
}

export function getSelectedTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  duration: string | undefined,
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
  } else if (duration && timeRangeSummary.min) {
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

const ExploreNameRegex = /\/explore\/((?:\w|-)+)/;
export function getExploreName(webOpenPath: string) {
  const matches = ExploreNameRegex.exec(webOpenPath);
  if (!matches || matches.length < 1) return "";
  return matches[1];
}

export async function convertQueryFilterToToplistQuery(
  instanceId: string,
  metricsView: string,
  req: V1MetricsViewAggregationRequest,
  dimension: string,
) {
  const params = <QueryServiceMetricsViewAggregationBody>{
    ...req,
  };
  const toplist = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewAggregationQueryKey(
      instanceId,
      metricsView,
      params,
    ),
    queryFn: () =>
      queryServiceMetricsViewAggregation(instanceId, metricsView, params),
  });
  return createInExpression(
    dimension,
    toplist.data?.map((d) => d[dimension]) ?? [],
  );
}

export async function getExplorePageUrlSearchParams(
  exploreName: string,
  exploreState: MetricsExplorerEntity,
): Promise<URLSearchParams> {
  const instanceId = get(runtime).instanceId;
  const { explore, metricsView } = await queryClient.fetchQuery({
    queryFn: ({ signal }) =>
      runtimeServiceGetExplore(
        instanceId,
        {
          name: exploreName,
        },
        signal,
      ),
    queryKey: getRuntimeServiceGetExploreQueryKey(instanceId, {
      name: exploreName,
    }),
    // this loader function is run for every param change in url.
    // so to avoid re-fetching explore everytime we set this so that it hits cache.
    staleTime: Infinity,
  });

  const metricsViewSpec = metricsView?.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore?.explore?.state?.validSpec ?? {};
  const metricsViewName = exploreSpec.metricsView;

  let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined;
  if (
    metricsView?.metricsView?.state?.validSpec?.timeDimension &&
    metricsViewName
  ) {
    fullTimeRange = await queryClient.fetchQuery({
      queryFn: () =>
        queryServiceMetricsViewTimeRange(instanceId, metricsViewName, {}),
      queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
        instanceId,
        metricsViewName,
        {},
      ),
      staleTime: Infinity,
      cacheTime: Infinity,
    });
  }

  const searchParams = convertExploreStateToURLSearchParams(
    exploreState,
    exploreSpec,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      fullTimeRange?.timeRangeSummary,
      exploreState,
    ),
    getDefaultExplorePreset(exploreSpec, fullTimeRange),
  );

  return searchParams;
}
