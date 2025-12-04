import { page } from "$app/stores";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import {
  type MapQueryRequest,
  type MapQueryStateOptions,
  mapQueryToDashboard,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, readable } from "svelte/store";

export type MapExploreUrlContext = {
  instanceId: string;
  organization: string;
  project: string;
  token?: string;
};

/**
 * Returns a store of explore URL with state filled in based on a query and queryRequestProperties.
 * Takes {@link MapQueryRequest} and {@link MapQueryStateOptions} that is directly passed to {@link mapQueryToDashboard}
 * Also takes {@link MapExploreUrlContext} to finally build the url.
 */
export function getMappedExploreUrl(
  req: MapQueryRequest, // Request object passed directly to mapQueryToDashboard
  opts: MapQueryStateOptions, // Map options passed directly to mapQueryToDashboard
  { instanceId, organization, project, token }: MapExploreUrlContext,
) {
  if (!req.queryArgsJson) return readable("");
  const queryRequestProperties = JSON.parse(req.queryArgsJson);
  const metricsViewName: string | undefined =
    queryRequestProperties.metricsView ||
    queryRequestProperties.metricsViewName;
  if (!metricsViewName) return readable("");

  return derived(
    [
      useExploreValidSpec(instanceId, req.exploreName, undefined, queryClient),
      useMetricsViewTimeRange(
        instanceId,
        metricsViewName,
        undefined,
        queryClient,
      ),
      mapQueryToDashboard(req, opts),
      page,
    ],
    ([validSpecResp, timeRangeSummaryResp, dashboardState, pageState]) => {
      const url = new URL(pageState.url);
      if (token) {
        url.pathname = `/${organization}/${project}/-/share/${token}/explore/${req.exploreName}`;
      } else {
        url.pathname = `/${organization}/${project}/explore/${req.exploreName}`;
      }

      if (!dashboardState?.data?.exploreState || !validSpecResp.data) {
        return url.toString();
      }

      const metricsViewSpec = validSpecResp.data.metricsView ?? {};
      const exploreSpec = validSpecResp.data.explore ?? {};

      const searchParams = convertPartialExploreStateToUrlParams(
        exploreSpec,
        dashboardState.data.exploreState,
        getTimeControlState(
          metricsViewSpec,
          exploreSpec,
          timeRangeSummaryResp.data?.timeRangeSummary,
          dashboardState.data.exploreState,
        ),
      );
      url.search = searchParams.toString();

      return url.toString();
    },
  );
}
