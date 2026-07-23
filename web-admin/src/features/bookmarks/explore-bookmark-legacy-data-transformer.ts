import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto.ts";
import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function createExploreBookmarkLegacyDataTransformer(
  client: RuntimeClient,
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(client, exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(client, exploreNameStore),
  );

  return derived(
    [validSpecQuery, timeRangeQuery],
    ([validSpecResp, timeRangeResp]) => {
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;

      return (data: string) =>
        exploreBookmarkDataTransformer({
          data,
          metricsViewSpec,
          exploreSpec,
          timeRangeSummary,
        });
    },
  );
}

export function exploreBookmarkDataTransformer({
  data,
  metricsViewSpec,
  exploreSpec,
  timeRangeSummary,
}: {
  data: string;
  metricsViewSpec: V1MetricsViewSpec;
  exploreSpec: V1ExploreSpec;
  timeRangeSummary: V1TimeRangeSummary | undefined;
}) {
  const exploreStateFromBookmark = getDashboardStateFromUrl(
    data,
    metricsViewSpec,
    exploreSpec,
  );

  // We need to check if the bookmark's url is equal to current url or not to show an "active" state.
  // To avoid calculating it everytime we directly convert it to final url.
  const searchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    exploreStateFromBookmark,
    getTimeControlState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
      exploreStateFromBookmark,
    ),
  );

  // Legacy protobuf bookmarks predate these two Explore settings. Complete
  // bookmarks inherit them, while filter-only bookmarks must remain partial.
  const filterOnlyKeys = new Set([
    ExploreStateURLParams.Filters,
    ExploreStateURLParams.TimeRange,
    ExploreStateURLParams.TimeGrain,
  ]);
  const isFilterOnly = [...searchParams.keys()].every((key) =>
    filterOnlyKeys.has(key as ExploreStateURLParams),
  );
  if (!isFilterOnly) {
    const defaultParams = getRillDefaultExploreUrlParams(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
    );
    for (const key of [
      ExploreStateURLParams.DynamicYAxisScale,
      ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
    ]) {
      const value = defaultParams.get(key);
      if (value !== null && !searchParams.has(key))
        searchParams.set(key, value);
    }
  }

  return "?" + searchParams.toString();
}
