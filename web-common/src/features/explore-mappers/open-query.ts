import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { getUrlForExplore } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  runtimeServiceListResources,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { error, redirect } from "@sveltejs/kit";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { createLinkError } from "@rilldata/web-common/features/explore-mappers/explore-validation.ts";
import { ExploreLinkErrorType } from "@rilldata/web-common/features/explore-mappers/types.ts";

export async function openQuery({
  query,
  organization,
  project,
  client,
}: {
  query: MetricsResolverQuery;
  client: RuntimeClient;
  organization?: string;
  project?: string;
}) {
  let exploreURL: string;

  try {
    // Extract metrics view name (now type-safe)
    const metricsViewName = query.metrics_view;
    if (!metricsViewName) {
      throw new Error("metrics_view is required in query");
    }

    // Find an explore dashboard that uses this metrics view
    const exploreName = await findExploreForMetricsView(
      client,
      metricsViewName,
    );

    const { metricsViewSpec, exploreSpec } = await getExploreSpecs(
      client,
      exploreName,
    );

    // Convert query to ExploreState
    const exploreState = mapMetricsResolverQueryToDashboard(
      metricsViewSpec,
      exploreSpec,
      query,
    );

    // Generate the final explore URL
    exploreURL = await generateExploreLink(
      client,
      exploreState,
      metricsViewSpec,
      exploreSpec,
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

async function getExploreSpecs(client: RuntimeClient, exploreName: string) {
  const request = { name: exploreName };
  const getExploreResponse = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetExploreQueryKey(client.instanceId, request),
    queryFn: ({ signal }) =>
      runtimeServiceGetExplore(client, request, { signal }),
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

  return { metricsViewSpec, exploreSpec };
}

/**
 * Generates the explore page URL with proper search parameters
 */
async function generateExploreLink(
  client: RuntimeClient,
  exploreState: Partial<ExploreState>,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreName: string,
  organization?: string | undefined,
  project?: string | undefined,
): Promise<string> {
  try {
    const url = getUrlForExplore(exploreName, organization, project);

    const metricsViewName = exploreSpec.metricsView;
    let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined;
    if (metricsViewSpec.timeDimension && metricsViewName) {
      const request = { metricsViewName };
      fullTimeRange = await queryClient.fetchQuery({
        queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
          client.instanceId,
          request,
        ),
        queryFn: ({ signal }) =>
          queryServiceMetricsViewTimeRange(client, request, { signal }),
        staleTime: Infinity,
        gcTime: Infinity,
      });
    }

    const searchParams = convertPartialExploreStateToUrlParams(
      exploreSpec,
      metricsViewSpec,
      exploreState,
      getTimeControlState(
        metricsViewSpec,
        exploreSpec,
        fullTimeRange?.timeRangeSummary,
        exploreState,
      ),
    );

    searchParams.forEach((value, key) => {
      url.searchParams.set(key, value);
    });

    return url.toString();
  } catch (error) {
    throw createLinkError(
      ExploreLinkErrorType.TRANSFORMATION_ERROR,
      `Failed to generate explore link: ${error.message}`,
      error,
    );
  }
}
