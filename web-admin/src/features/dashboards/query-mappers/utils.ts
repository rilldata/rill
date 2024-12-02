import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createInExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import {
  getQueryServiceMetricsViewAggregationQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  queryServiceMetricsViewAggregation,
  type QueryServiceMetricsViewAggregationBody,
  runtimeServiceGetExplore,
  type V1Expression,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
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
      !reqComparisonTimeRange.isoOffset &&
      reqComparisonTimeRange.isoDuration
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

export async function convertExprToToplist(
  queryClient: QueryClient,
  instanceId: string,
  metricsView: string,
  dimensionName: string,
  measureName: string,
  timeRange: V1TimeRange | undefined,
  comparisonTimeRange: V1TimeRange | undefined,
  executionTime: string,
  where: V1Expression | undefined,
  having: V1Expression,
) {
  let hasPercentOfTotals = false;
  forEachIdentifier(having, (_, ident) => {
    if (ident?.endsWith(ComparisonPercentOfTotal)) {
      hasPercentOfTotals = true;
    }
  });

  const toplistBody: QueryServiceMetricsViewAggregationBody = {
    measures: [
      {
        name: measureName,
      },
      ...(comparisonTimeRange
        ? [
            {
              name: measureName + ComparisonDeltaAbsoluteSuffix,
              comparisonDelta: { measure: measureName },
            },
            {
              name: measureName + ComparisonDeltaRelativeSuffix,
              comparisonRatio: { measure: measureName },
            },
          ]
        : []),
      ...(hasPercentOfTotals
        ? [
            {
              name: measureName + ComparisonPercentOfTotal,
              percentOfTotal: { measure: measureName },
            },
          ]
        : []),
    ],
    dimensions: [{ name: dimensionName }],
    ...(timeRange
      ? {
          timeRange: {
            ...timeRange,
            end: executionTime,
          },
        }
      : {}),
    ...(comparisonTimeRange
      ? {
          comparisonTimeRange: {
            ...comparisonTimeRange,
            end: executionTime,
          },
        }
      : {}),
    where,
    having,
    sort: [
      {
        name: measureName,
        desc: false,
      },
    ],
    limit: "250",
  };
  const toplist = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewAggregationQueryKey(
      instanceId,
      metricsView,
      toplistBody,
    ),
    queryFn: () =>
      queryServiceMetricsViewAggregation(instanceId, metricsView, toplistBody),
  });
  if (!toplist.data) {
    return undefined;
  }
  return createInExpression(
    dimensionName,
    toplist.data.map((t) => t[dimensionName]),
  );
}

const ExploreNameRegex = /\/explore\/((?:\w|-)+)/;
export function getExploreName(webOpenPath: string) {
  const matches = ExploreNameRegex.exec(webOpenPath);
  if (!matches || matches.length < 1) return "";
  return matches[1];
}

export async function getExplorePageUrl(
  curPageUrl: URL,
  organization: string,
  project: string,
  exploreName: string,
  dashboard: MetricsExplorerEntity,
) {
  const instanceId = get(runtime).instanceId;
  const { explore } = await queryClient.fetchQuery({
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

  const url = new URL(`${curPageUrl.protocol}//${curPageUrl.host}`);
  url.pathname = `/${organization}/${project}/explore/${exploreName}`;

  const exploreSpec = explore?.explore?.state?.validSpec;
  const searchParamsFromMetrics = convertExploreStateToURLSearchParams(
    dashboard,
    exploreSpec ?? {},
    exploreSpec?.defaultPreset ?? {},
  );
  mergeSearchParams(searchParamsFromMetrics, url.searchParams);
  return url.toString();
}
