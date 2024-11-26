import type { QueryFunction } from "@rilldata/svelte-query";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";

export async function fetchExploreSpec(
  instanceId: string,
  exploreName: string,
) {
  const queryParams = {
    name: exploreName,
  };
  const queryKey = getRuntimeServiceGetExploreQueryKey(instanceId, queryParams);
  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetExplore>>
  > = ({ signal }) => runtimeServiceGetExplore(instanceId, queryParams, signal);

  const response = await queryClient.fetchQuery({
    queryFn: queryFunction,
    queryKey,
    staleTime: Infinity,
  });

  const exploreResource = response.explore;
  const metricsViewResource = response.metricsView;

  if (!exploreResource?.explore) {
    throw error(404, "Explore not found");
  }
  if (!metricsViewResource?.metricsView) {
    throw error(404, "Metrics view not found");
  }

  let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined = undefined;
  const metricsViewName = exploreResource.explore.state?.validSpec?.metricsView;
  if (
    metricsViewResource.metricsView.state?.validSpec?.timeDimension &&
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
    });
  }

  const defaultExplorePreset = getBasePreset(
    exploreResource.explore.state?.validSpec ?? {},
    getLocalUserPreferencesState(exploreName),
    fullTimeRange,
  );

  return {
    explore: exploreResource,
    metricsView: metricsViewResource,
    defaultExplorePreset,
  };
}
