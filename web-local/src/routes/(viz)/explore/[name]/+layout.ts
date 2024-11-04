import type { QueryFunction } from "@rilldata/svelte-query";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params, depends }) => {
  const { instanceId } = get(runtime);

  const exploreName = params.name;

  depends(exploreName, "explore");

  const queryParams = {
    name: exploreName,
  };

  const queryKey = getRuntimeServiceGetExploreQueryKey(instanceId, queryParams);

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetExplore>>
  > = ({ signal }) => runtimeServiceGetExplore(instanceId, queryParams, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const exploreResource = response.explore;
    const metricsViewResource = response.metricsView;

    if (!exploreResource?.explore) {
      throw error(404, "Explore not found");
    }
    if (!metricsViewResource?.metricsView) {
      throw error(404, "Metrics view not found");
    }

    const basePreset = getBasePreset(
      exploreResource.explore.state?.validSpec ?? {},
      getLocalUserPreferencesState(exploreName),
    );

    return {
      explore: exploreResource,
      metricsView: metricsViewResource,
      basePreset,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Explore not found");
  }
};
