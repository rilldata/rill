import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import { correctExploreState } from "@rilldata/web-common/features/dashboards/stores/correct-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { cleanEmbedUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { getCleanedUrlParamsForGoto } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceMetricsViewTimeRangeQueryOptions,
  getRuntimeServiceGetExploreQueryOptions,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export type ValidatedExploreUrl = {
  // The cleaned and canonicalized url params for the given state.
  // Invalid params are dropped, so this is always safe to apply even when `errors` is non-empty.
  url: URLSearchParams;
  errors: Error[];
};

/**
 * Converts a set of url params into the canonical explore url, collecting any validation errors.
 *
 * This reuses the same conversion parts as DashboardStateSync.handleURLChange so that the returned
 * url matches what the dashboard would end up on: url params are validated against the explore and
 * metrics view specs, back-filled with rill opinionated defaults, and then converted back to url
 * params with defaults removed.
 *
 * Unlike handleURLChange it does not mutate the explore store, touch session storage or navigate.
 * It also intentionally skips the session storage, most recent and bookmark sources so that the
 * resulting url is determined solely by the provided params (plus rill defaults), not the viewer's
 * local state.
 *
 * Errors are returned rather than thrown; the caller decides whether to apply the url regardless.
 */
export async function buildValidatedExploreUrl(
  client: RuntimeClient,
  exploreName: string,
  searchParams: URLSearchParams,
  currentUrl: URL,
): Promise<ValidatedExploreUrl> {
  const exploreResp = await queryClient.fetchQuery(
    getRuntimeServiceGetExploreQueryOptions(client, { name: exploreName }),
  );
  const metricsViewSpec =
    exploreResp.metricsView?.metricsView?.state?.validSpec ?? {};
  const exploreSpec = exploreResp.explore?.explore?.state?.validSpec ?? {};

  let timeRangeSummary: V1TimeRangeSummary | undefined;
  if (metricsViewSpec.timeDimension) {
    const timeRangeResp = await queryClient.fetchQuery(
      getQueryServiceMetricsViewTimeRangeQueryOptions(client, {
        metricsViewName: exploreSpec.metricsView ?? "",
      }),
    );
    timeRangeSummary = timeRangeResp.timeRangeSummary;
  }

  const cleanedParams = cleanEmbedUrlParams(searchParams);
  const { partialExploreState, errors } = convertURLSearchParamsToExploreState(
    cleanedParams,
    metricsViewSpec,
    exploreSpec,
    {},
  );

  // Back-fill rill opinionated defaults so that fields not present in the url (like time range)
  // resolve to a complete, valid state.
  const rillDefaultExploreState = getRillDefaultExploreState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const exploreState = cascadingExploreStateMerge([
    partialExploreState,
    rillDefaultExploreState,
  ]) as ExploreState;
  correctExploreState(metricsViewSpec, exploreState);

  const timeControlsState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    exploreState,
  );
  const defaultExploreUrlParams = getRillDefaultExploreUrlParams(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );

  const url = getCleanedUrlParamsForGoto(
    exploreSpec,
    metricsViewSpec,
    exploreState,
    timeControlsState,
    defaultExploreUrlParams,
    currentUrl,
  );

  return { url, errors };
}
