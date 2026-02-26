import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { resolveTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapV1TimeRangeToSelectedComparisonTimeRange,
  mapV1TimeRangeToSelectedTimeRange,
  PreviousCompleteRangeMap,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import {
  type ExploreLinkError,
  ExploreLinkErrorType,
} from "@rilldata/web-common/features/explore-mappers/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
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
  type V1ExploreSpec,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
// We are manually sending in duration, offset and round to grain for previous complete ranges.
// This is to map back that split
const PreviousCompleteRangeReverseMap: Record<string, TimeRangePreset> = {};
for (const preset in PreviousCompleteRangeMap) {
  const range: V1TimeRange = PreviousCompleteRangeMap[preset];
  PreviousCompleteRangeReverseMap[
    `${range.isoDuration}_${range.isoOffset}_${range.roundToGrain}`
  ] = preset as TimeRangePreset;
}

export async function fillTimeRange(
  client: RuntimeClient,
  exploreSpec: V1ExploreSpec,
  exploreState: ExploreState,
  reqTimeRange: V1TimeRange | undefined,
  reqComparisonTimeRange: V1TimeRange | undefined,
  timeRangeSummary: V1TimeRangeSummary,
  executionTime?: string,
) {
  const endTime =
    executionTime ?? timeRangeSummary.max ?? new Date().toISOString();

  if (reqTimeRange) {
    exploreState.selectedTimeRange = mapV1TimeRangeToSelectedTimeRange(
      reqTimeRange,
      timeRangeSummary,
      endTime,
    );
    if (
      exploreState.selectedTimeRange?.start &&
      exploreState.selectedTimeRange?.end &&
      executionTime
    ) {
      exploreState.selectedTimeRange.name = TimeRangePreset.CUSTOM;
    }
  }

  if (reqComparisonTimeRange) {
    if (
      (!reqComparisonTimeRange.isoOffset &&
        reqComparisonTimeRange.isoDuration) ||
      (reqComparisonTimeRange.isoOffset &&
        reqComparisonTimeRange.isoOffset === reqComparisonTimeRange.isoDuration)
    ) {
      exploreState.selectedComparisonTimeRange = {
        name: TimeComparisonOption.CONTIGUOUS,
        start: undefined as unknown as Date,
        end: undefined as unknown as Date,
      };
    } else {
      exploreState.selectedComparisonTimeRange =
        mapV1TimeRangeToSelectedComparisonTimeRange(
          reqComparisonTimeRange,
          timeRangeSummary,
          endTime,
        );
    }

    if (exploreState.selectedComparisonTimeRange) {
      exploreState.selectedComparisonTimeRange.interval =
        exploreState.selectedTimeRange?.interval;
    }
    exploreState.showTimeComparison = true;
  }

  // Resolve time range overriding ref to `executionTime` and set to custom.
  // This keeps the time range consistent regardless of when the link is opened.
  [exploreState.selectedTimeRange] = await resolveTimeRanges(
    client,
    exploreSpec,
    [exploreState.selectedTimeRange],
    exploreState.selectedTimezone,
    executionTime,
  );
  if (exploreState.selectedTimeRange) {
    exploreState.selectedTimeRange.name = TimeRangePreset.CUSTOM;
  }
}

const ExploreNameRegex = /\/explore\/((?:[\w-]|%[0-9A-Fa-f]{2})+)/;

export function getExploreName(webOpenPath: string) {
  const matches = ExploreNameRegex.exec(webOpenPath);

  if (!matches || matches.length < 1) return "";

  return decodeURIComponent(matches[1]);
}

export async function convertQueryFilterToToplistQuery(
  client: RuntimeClient,
  metricsView: string,
  req: V1MetricsViewAggregationRequest,
  dimension: string,
) {
  const params = {
    metricsViewName: metricsView,
    ...req,
  };
  const toplist = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewAggregationQueryKey(
      client.instanceId,
      params,
    ),
    queryFn: () => queryServiceMetricsViewAggregation(client, params),
  });
  return createInExpression(
    dimension,
    toplist.data?.map((d) => d[dimension]) ?? [],
  );
}

export async function getExplorePageUrlSearchParams(
  client: RuntimeClient,
  exploreName: string,
  exploreState: Partial<ExploreState>,
): Promise<URLSearchParams> {
  const { explore, metricsView } = await queryClient.fetchQuery({
    queryFn: ({ signal }) =>
      runtimeServiceGetExplore(
        client,
        {
          name: exploreName,
        },
        { signal },
      ),
    queryKey: getRuntimeServiceGetExploreQueryKey(client.instanceId, {
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
        queryServiceMetricsViewTimeRange(client, { metricsViewName }),
      queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(client.instanceId, {
        metricsViewName,
      }),
      staleTime: Infinity,
      gcTime: Infinity,
    });
  }

  // This is just for an initial redirect.
  // DashboardStateDataLoader will handle compression etc. during init
  // So no need to use getCleanedUrlParamsForGoto
  const searchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    exploreState,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      fullTimeRange?.timeRangeSummary,
      exploreState,
    ),
  );

  return searchParams;
}

/**
 * Sync method to get explore page url search params if metrics view time range is present in cache.
 * Else the map will have to happen in the `/-/open-query` route.
 * @param exploreState
 * @param metricsViewSpec
 * @param exploreSpec
 */
export function maybeGetExplorePageUrlSearchParams(
  instanceId: string,
  exploreState: Partial<ExploreState>,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  const metricsViewName = exploreSpec.metricsView ?? "";

  const metricsViewTimeRangeQueryKey =
    getQueryServiceMetricsViewTimeRangeQueryKey(instanceId, {
      metricsViewName,
    });

  // Get time range query from cache if present, else we will go to `/-/open-query` to fetch it.
  const queryResp = queryClient.getQueryData<V1MetricsViewTimeRangeResponse>(
    metricsViewTimeRangeQueryKey,
  );
  if (!queryResp) return null;

  // This is just for an initial redirect.
  // DashboardStateDataLoader will handle compression etc. during init
  // So no need to use getCleanedUrlParamsForGoto
  const searchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    exploreState,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      queryResp.timeRangeSummary,
      exploreState,
    ),
  );

  return searchParams;
}

/**
 * This method corrects the underscore naming to camel case.
 * This is the drawback of storing the request object as is.
 */
export function convertRequestKeysToCamelCase(
  req: Record<string, any>,
): Record<string, any> {
  const newReq: Record<string, any> = {};

  for (const key in req) {
    const newKey = key.replace(/_(\w)/g, (_, c: string) => c.toUpperCase());
    const val = req[key];
    if (val && typeof val === "object" && !("length" in val)) {
      newReq[newKey] = convertRequestKeysToCamelCase(val);
    } else {
      newReq[newKey] = val;
    }
  }

  return newReq;
}

export function getErrorMessage(error: ExploreLinkError): string {
  if (error.message) {
    return error.message;
  }
  switch (error.type) {
    case ExploreLinkErrorType.VALIDATION_ERROR:
      return "No compatible explore dashboard found for this component.";
    case ExploreLinkErrorType.PERMISSION_ERROR:
      return "You do not have permission to access the explore dashboard.";
    case ExploreLinkErrorType.NETWORK_ERROR:
      return "Failed to connect to the server. Please try again.";
    case ExploreLinkErrorType.TRANSFORMATION_ERROR:
    default:
      return "Unable to open explore dashboard. Please try again.";
  }
}
