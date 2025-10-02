import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceGetExplore,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * Apply metricsProps to the current dashboard state.
 * This function converts the metricsProps (which are in MetricsResolverQuery format)
 * to ExploreState and applies them to the dashboard.
 */
export async function applyMetricsPropsToDashboard(
  metricsProps: MetricsResolverQuery,
  exploreName: string,
): Promise<{ partialExploreState: Partial<ExploreState>; metricsViewSpec: any }> {
  const instanceId = get(runtime).instanceId;

  // Get explore and metrics view specs
  const getExploreResponse = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetExploreQueryKey(instanceId, {
      name: exploreName,
    }),
    queryFn: ({ signal }) =>
      runtimeServiceGetExplore(
        instanceId,
        {
          name: exploreName,
        },
        signal,
      ),
  });
  const exploreResource = getExploreResponse.explore;
  const metricsViewResource = getExploreResponse.metricsView;

  if (!exploreResource?.explore?.state?.validSpec) {
    throw new Error("Could not load explore specification");
  }

  if (!metricsViewResource?.metricsView?.state?.validSpec) {
    throw new Error("Could not load metrics view specification");
  }

  const metricsViewSpec = metricsViewResource?.metricsView?.state?.validSpec;
  const exploreSpec = exploreResource.explore.state.validSpec;

  // Debug logging
  console.log("Debug - metricsViewSpec:", metricsViewSpec);
  console.log("Debug - exploreSpec:", exploreSpec);
  console.log("Debug - metricsProps:", metricsProps);

  if (!metricsViewSpec) {
    throw new Error("Metrics view specification is undefined");
  }

  if (!exploreSpec) {
    throw new Error("Explore specification is undefined");
  }

  // Convert metricsProps to ExploreState
  const partialExploreState: Partial<ExploreState> =
    mapMetricsResolverQueryToDashboard(metricsViewSpec, exploreSpec, metricsProps);

  return { partialExploreState, metricsViewSpec };
}

/**
 * Find an explore dashboard that uses the given metrics view.
 * This mirrors the backend's findExploreForMetricsView logic.
 */
export async function findExploreForMetricsView(
  metricsViewName: string,
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // List all explore resources
  const exploreResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: ResourceKind.Explore,
    }),
    queryFn: ({ signal }) =>
      runtimeServiceListResources(
        instanceId,
        { kind: ResourceKind.Explore },
        signal,
      ),
  });

  // Look for an explore that references this metrics view
  if (exploreResources.resources) {
    for (const resource of exploreResources.resources) {
      if (resource.explore?.state?.validSpec?.metricsView === metricsViewName) {
        return resource.meta?.name?.name || "";
      }
    }
  }

  // If no explore found, throw an error
  throw new Error(
    `No explore dashboard found for metrics view: ${metricsViewName}`,
  );
}
