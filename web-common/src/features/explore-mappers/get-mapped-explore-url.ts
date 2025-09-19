import { page } from "$app/stores";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import {
  type MapQueryRequest,
  mapQueryToDashboard,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived } from "svelte/store";

export function getMappedExploreUrl(
  req: MapQueryRequest,
  {
    instanceId,
    organization,
    project,
    token,
  }: {
    instanceId: string;
    organization: string;
    project: string;
    token?: string;
  },
) {
  const queryRequestProperties = JSON.parse(req.queryArgsJson ?? "{}");
  const metricsViewName =
    queryRequestProperties.metricsView ??
    queryRequestProperties.metricsViewName ??
    "";

  return derived(
    [
      useExploreValidSpec(instanceId, req.exploreName, undefined, queryClient),
      useMetricsViewTimeRange(
        instanceId,
        metricsViewName,
        undefined,
        queryClient,
      ),
      mapQueryToDashboard(req),
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
