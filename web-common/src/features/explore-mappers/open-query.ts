import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { getUrlForExplore } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1ListResourcesResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { error, redirect } from "@sveltejs/kit";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import httpClient from "@rilldata/web-common/runtime-client/http-client.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { createLinkError } from "@rilldata/web-common/features/explore-mappers/explore-validation.ts";
import { ExploreLinkErrorType } from "@rilldata/web-common/features/explore-mappers/types.ts";

export async function openQuery({
  url,
  organization,
  project,
  runtime,
}: {
  url: URL;
  runtime: Runtime;
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
    const exploreName = await findExploreForMetricsView(
      runtime,
      metricsViewName,
    );

    const { metricsViewSpec, exploreSpec } = await getExploreSpecs(
      runtime,
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
      runtime,
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
  runtime: Runtime,
  metricsViewName: string,
): Promise<string> {
  // List all explore resources
  const exploreResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(runtime.instanceId, {
      kind: ResourceKind.Explore,
    }),
    queryFn: ({ signal }) =>
      httpClient<V1ListResourcesResponse>({
        url: `/v1/instances/${runtime.instanceId}/resources`,
        method: "GET",
        params: { kind: ResourceKind.Explore },
        signal,
        baseUrl: runtime.host,
        headers: runtime.jwt
          ? {
              Authorization: `Bearer ${runtime.jwt?.token}`,
            }
          : undefined,
      }),
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

async function getExploreSpecs(runtime: Runtime, exploreName: string) {
  // Get explore and metrics view specs
  const getExploreResponse = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetExploreQueryKey(runtime.instanceId, {
      name: exploreName,
    }),
    queryFn: ({ signal }) =>
      httpClient<V1GetExploreResponse>({
        url: `/v1/instances/${runtime.instanceId}/resources/explore`,
        method: "GET",
        params: { name: exploreName },
        signal,
        baseUrl: runtime.host,
        headers: runtime.jwt
          ? {
              Authorization: `Bearer ${runtime.jwt?.token}`,
            }
          : undefined,
      }),
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
  runtime: Runtime,
  exploreState: Partial<ExploreState>,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreName: string,
  organization?: string | undefined,
  project?: string | undefined,
): Promise<string> {
  try {
    // Build base URL
    const url = getUrlForExplore(exploreName, organization, project);

    const metricsViewName = exploreSpec.metricsView;
    let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined;
    if (metricsViewSpec.timeDimension && metricsViewName) {
      fullTimeRange = await queryClient.fetchQuery({
        queryFn: ({ signal }) =>
          httpClient<V1MetricsViewTimeRangeResponse>({
            url: `/v1/instances/${runtime.instanceId}/queries/metrics-views/${metricsViewName}/time-range-summary`,
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              ...(runtime.jwt
                ? {
                    Authorization: `Bearer ${runtime.jwt?.token}`,
                  }
                : {}),
            },
            data: {},
            signal,
            baseUrl: runtime.host,
          }),
        queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
          runtime.instanceId,
          metricsViewName,
          {},
        ),
        staleTime: Infinity,
        gcTime: Infinity,
      });
    }

    // This is just for an initial redirect.
    // DashboardStateDataLoader will handle compression etc. during init
    // So no need to use getCleanedUrlParamsForGoto
    const searchParams = convertPartialExploreStateToUrlParams(
      exploreSpec,
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
