import { page } from "$app/stores";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import {
  type AIResolverProps,
  mapAIResolverPropsToMetricsResolverQuery,
} from "@rilldata/web-common/features/explore-mappers/map-ai-resolver-props-to-metrics-resolver-query.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import {
  type MapQueryRequest,
  type MapQueryStateOptions,
  mapQueryToDashboard,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, readable, type Readable } from "svelte/store";
import type { V1MetricsViewTimeRangeResponse } from "@rilldata/web-common/runtime-client";

export type MapExploreUrlContext = {
  client: RuntimeClient;
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
  { client, organization, project, token }: MapExploreUrlContext,
) {
  if (!req.queryArgsJson) return readable("");
  const queryRequestProperties = JSON.parse(req.queryArgsJson);
  const metricsViewName: string | undefined =
    queryRequestProperties.metricsView ||
    queryRequestProperties.metricsViewName;
  if (!metricsViewName) return readable("");

  return derived(
    [
      useExploreValidSpec(client, req.exploreName, undefined, queryClient),
      useMetricsViewTimeRange(client, metricsViewName, undefined, queryClient),
      mapQueryToDashboard(client, req, opts),
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
        metricsViewSpec,
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

/**
 * Returns a store of the explore URL for an AI report, with the report's scope (dimensions, measures, time ranges and filter)
 * applied as explore state. Reports that only name an explore get the plain explore URL.
 */
export function getMappedAIExploreUrl(
  props: AIResolverProps,
  exploreName: string,
  { client, organization, project, token }: MapExploreUrlContext,
) {
  if (!exploreName) return readable("");

  const validSpec = useExploreValidSpec(
    client,
    exploreName,
    undefined,
    queryClient,
  );
  // The metrics view is only known once the explore's spec has loaded.
  const timeRangeSummary: Readable<V1MetricsViewTimeRangeResponse | undefined> =
    derived(validSpec, (validSpecResp, set) => {
      const metricsViewName = validSpecResp.data?.explore?.metricsView;
      if (!metricsViewName) {
        set(undefined);
        return;
      }
      return useMetricsViewTimeRange(
        client,
        metricsViewName,
        undefined,
        queryClient,
      ).subscribe((resp) => set(resp.data));
    });

  return derived(
    [validSpec, timeRangeSummary, page],
    ([validSpecResp, timeRangeSummaryResp, pageState]) => {
      const url = new URL(pageState.url);
      if (token) {
        url.pathname = `/${organization}/${project}/-/share/${token}/explore/${exploreName}`;
      } else {
        url.pathname = `/${organization}/${project}/explore/${exploreName}`;
      }
      url.search = "";

      const metricsViewSpec = validSpecResp.data?.metricsView;
      const exploreSpec = validSpecResp.data?.explore;
      if (!metricsViewSpec || !exploreSpec?.metricsView) {
        return url.toString();
      }

      const partialExploreState = mapMetricsResolverQueryToDashboard(
        metricsViewSpec,
        exploreSpec,
        {
          query: mapAIResolverPropsToMetricsResolverQuery(
            props,
            exploreSpec.metricsView,
          ),
        },
      );
      const searchParams = convertPartialExploreStateToUrlParams(
        exploreSpec,
        metricsViewSpec,
        partialExploreState,
        getTimeControlState(
          metricsViewSpec,
          exploreSpec,
          timeRangeSummaryResp?.timeRangeSummary,
          partialExploreState,
        ),
      );
      url.search = searchParams.toString();

      return url.toString();
    },
  );
}
