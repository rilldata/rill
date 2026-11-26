import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { getUrlForExplore } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { type MetricsResolverQueryMapperArgs } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryOptions,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { error, redirect } from "@sveltejs/kit";
import { convertMetricsResolverQueryToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-metrics-resolver-query-to-url-params.ts";

export async function openQuery({
  mapArgs,
  organization,
  project,
  client,
}: {
  mapArgs: MetricsResolverQueryMapperArgs;
  client: RuntimeClient;
  organization?: string;
  project?: string;
}) {
  let exploreURL: string;

  try {
    // Extract metrics view name (now type-safe)
    const metricsViewName = mapArgs.query.metrics_view;
    if (!metricsViewName) {
      throw new Error("metrics_view is required in query");
    }

    // Find an explore dashboard that uses this metrics view
    const exploreName = await findExploreForMetricsView(
      client,
      metricsViewName,
    );

    // A filter on the time dimension is moved to the explore's time range.
    const metricsViewResp = await queryClient.fetchQuery(
      getRuntimeServiceGetResourceQueryOptions(client, {
        name: { kind: ResourceKind.MetricsView, name: metricsViewName },
      }),
    );
    const timeDimension =
      metricsViewResp.resource?.metricsView?.state?.validSpec?.timeDimension;

    // Generate the final explore URL
    exploreURL = generateExploreLink(
      mapArgs,
      exploreName,
      timeDimension,
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
  client: RuntimeClient,
  metricsViewName: string,
): Promise<string> {
  const request = { kind: ResourceKind.Explore };
  const exploreResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(
      client.instanceId,
      request,
    ),
    queryFn: ({ signal }) =>
      runtimeServiceListResources(client, request, { signal }),
  });

  if (exploreResources.resources) {
    for (const resource of exploreResources.resources) {
      if (resource.explore?.state?.validSpec?.metricsView === metricsViewName) {
        return resource.meta?.name?.name || "";
      }
    }
  }

  throw new Error(
    `No explore dashboard found for metrics view: ${metricsViewName}`,
  );
}

/**
 * Generates the explore page URL with proper search parameters
 */
function generateExploreLink(
  args: MetricsResolverQueryMapperArgs,
  exploreName: string,
  timeDimension: string | undefined,
  organization?: string | undefined,
  project?: string | undefined,
): string {
  const url = getUrlForExplore(exploreName, organization, project);

  const searchParams = convertMetricsResolverQueryToUrlParams(
    args,
    timeDimension,
  );

  searchParams.forEach((value, key) => {
    url.searchParams.set(key, value);
  });

  return url.toString();
}
