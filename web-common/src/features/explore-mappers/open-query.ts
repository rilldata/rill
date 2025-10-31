import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { generateExploreLink } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export async function openQuery({
  url,
  organization,
  project,
}: {
  url: URL;
  organization?: string;
  project?: string;
}) {
  let exploreURL: string;

  try {
    const queryParams = url.searchParams;

    // Get the JSON-encoded query parameters
    const queryJSON = queryParams.get("query");
    if (!queryJSON) {
      throw new Error("query parameter is required");
    }

    // Parse and validate the query with proper type safety
    let query: MetricsResolverQuery;
    try {
      query = JSON.parse(queryJSON) as MetricsResolverQuery;
    } catch (e) {
      throw new Error(`Invalid query: ${e.message}`);
    }

    // Extract metrics view name (now type-safe)
    const metricsViewName = query.metrics_view;
    if (!metricsViewName) {
      throw new Error("metrics_view is required in query");
    }

    // Find an explore dashboard that uses this metrics view
    const exploreName = await findExploreForMetricsView(metricsViewName);

    // Convert query to ExploreState
    const exploreState = await convertQueryToExploreState(query, exploreName);

    // Generate the final explore URL
    exploreURL = await generateExploreLink(
      exploreState,
      exploreName,
      organization,
      project,
    );
  } catch (e) {
    console.error("Failed to process open-query:", e);

    // Use SvelteKit's error handling instead of manual redirect
    throw error(
      400,
      e.message || "Unable to open a dashboard that represents this query",
    );
  }

  // Redirect outside the try/catch since redirect() throws internally
  redirect(302, exploreURL);
}

/**
 * Find an explore dashboard that uses the given metrics view.
 * This mirrors the backend's findExploreForMetricsView logic.
 * TODO: try to find an explore that has as many measures/dimensions in the query
 */
async function findExploreForMetricsView(
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

/**
 * Convert Query directly to ExploreState without going through URL parameters.
 */
async function convertQueryToExploreState(
  query: MetricsResolverQuery,
  exploreName: string,
): Promise<Partial<ExploreState>> {
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
  const metricsViewName = exploreSpec.metricsView ?? "";

  const metricsViewTimeRangeResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
      instanceId,
      metricsViewName,
      {},
    ),
    queryFn: () =>
      queryServiceMetricsViewTimeRange(instanceId, metricsViewName, {}),
  });

  const partialExploreState: Partial<ExploreState> =
    mapMetricsResolverQueryToDashboard(
      metricsViewSpec,
      exploreSpec,
      metricsViewTimeRangeResp.timeRangeSummary,
      query,
    );

  return partialExploreState;
}
