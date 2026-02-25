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
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { createLinkError } from "@rilldata/web-common/features/explore-mappers/explore-validation.ts";
import { ExploreLinkErrorType } from "@rilldata/web-common/features/explore-mappers/types.ts";

interface RuntimeInfo {
  host: string;
  instanceId: string;
  jwt?: { token: string } | undefined;
}

async function runtimeFetch<T>(
  runtime: RuntimeInfo,
  path: string,
  opts?: { method?: string; body?: unknown; signal?: AbortSignal },
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (runtime.jwt) {
    headers["Authorization"] = `Bearer ${runtime.jwt.token}`;
  }
  const resp = await fetch(`${runtime.host}${path}`, {
    method: opts?.method ?? "GET",
    headers,
    ...(opts?.body !== undefined ? { body: JSON.stringify(opts.body) } : {}),
    signal: opts?.signal,
  });
  if (!resp.ok) {
    const data = await resp.json().catch(() => ({}));
    throw { response: { status: resp.status, data } };
  }
  return (await resp.json()) as T;
}

export async function openQuery({
  query,
  organization,
  project,
  runtime,
}: {
  query: MetricsResolverQuery;
  runtime: RuntimeInfo;
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
  runtime: RuntimeInfo,
  metricsViewName: string,
): Promise<string> {
  const exploreResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(runtime.instanceId, {
      kind: ResourceKind.Explore,
    }),
    queryFn: ({ signal }) =>
      runtimeFetch<V1ListResourcesResponse>(
        runtime,
        `/v1/instances/${runtime.instanceId}/resources?kind=${ResourceKind.Explore}`,
        { signal },
      ),
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

async function getExploreSpecs(runtime: RuntimeInfo, exploreName: string) {
  const getExploreResponse = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetExploreQueryKey(runtime.instanceId, {
      name: exploreName,
    }),
    queryFn: ({ signal }) =>
      runtimeFetch<V1GetExploreResponse>(
        runtime,
        `/v1/instances/${runtime.instanceId}/resources/explore?name=${encodeURIComponent(exploreName)}`,
        { signal },
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

  return { metricsViewSpec, exploreSpec };
}

/**
 * Generates the explore page URL with proper search parameters
 */
async function generateExploreLink(
  runtime: RuntimeInfo,
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
      fullTimeRange = await queryClient.fetchQuery({
        queryFn: ({ signal }) =>
          runtimeFetch<V1MetricsViewTimeRangeResponse>(
            runtime,
            `/v1/instances/${runtime.instanceId}/queries/metrics-views/${encodeURIComponent(metricsViewName)}/time-range-summary`,
            { method: "POST", body: {}, signal },
          ),
        queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
          runtime.instanceId,
          metricsViewName,
          {},
        ),
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
