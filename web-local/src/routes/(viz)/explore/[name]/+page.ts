import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getMetricsExplorerFromUrl } from "@rilldata/web-common/features/dashboards/url-state/fromUrl";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import type { QueryFunction } from "@tanstack/svelte-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export const load = async ({ params, depends, url }) => {
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

    let partialMetrics: Partial<MetricsExplorerEntity> = {};
    if (
      metricsViewResource.metricsView.state?.validSpec &&
      exploreResource.explore.state?.validSpec &&
      url
    ) {
      const { entity, errors } = getMetricsExplorerFromUrl(
        url.searchParams,
        metricsViewResource.metricsView.state.validSpec,
        exploreResource.explore.state.validSpec,
        exploreResource.explore.state.validSpec.defaultPreset ?? {},
      );
      partialMetrics = entity;
      if (errors.length) console.log(errors); // TODO
    }

    return {
      explore: exploreResource,
      metricsView: metricsViewResource,
      partialMetrics,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Explore not found");
  }
};
