import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type {
  MapQueryRequest,
  MapQueryStateOptions,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import type { QueryRequests } from "@rilldata/web-common/features/explore-mappers/types.ts";
import {
  convertRequestKeysToCamelCase,
  type DashboardPathContext,
  getDashboardPath,
} from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { getUrlFromAggregationRequest } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-url-from-aggregation-request.ts";
import { getUrlFromComparisonRequest } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-url-from-comparison-request.ts";
import { mapMetricsResolverQueryToUrlSearch } from "@rilldata/web-common/features/explore-mappers/url-mappers/map-metrics-resolver-query-to-url-search.ts";
import {
  type AIResolverProps,
  mapAIResolverPropsToMetricsResolverQuery,
} from "@rilldata/web-common/features/explore-mappers/map-ai-resolver-props-to-metrics-resolver-query.ts";
import type {
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";

/**
 * What an explore url is mapped from:
 * the query of an alert or a report, or the scope of an AI report.
 */
export type MappedExploreUrlSource =
  | {
      type: "query";
      request: MapQueryRequest;
      options: MapQueryStateOptions;
    }
  | {
      type: "ai";
      exploreName: string;
      props: AIResolverProps;
    };

/**
 * Query options for the explore url of an alert or a report.
 * Mapping can fetch data (resolving time ranges, toplists for comparison measure filters), so it runs as a query:
 * callers get loading and error states, and results are cached across components.
 *
 * The specs are not fetched here; the url is validated when the explore is built from it.
 */
export function getMappedExploreUrlQueryOptions(
  client: RuntimeClient,
  source: MappedExploreUrlSource,
  ctx: DashboardPathContext,
) {
  return {
    queryKey: ["mapped-explore-url", client.instanceId, source, ctx] as const,
    queryFn: () => getMappedExploreUrl(client, source, ctx),
    enabled: canMapExploreUrl(source),
    // Relative time ranges are resolved to absolute ones, so a cached url stays valid.
    staleTime: Infinity,
  };
}

/**
 * Returns the path and search of the explore that shows the source.
 */
export async function getMappedExploreUrl(
  client: RuntimeClient,
  source: MappedExploreUrlSource,
  ctx: DashboardPathContext,
) {
  if (source.type === "ai") {
    const path = getDashboardPath(ctx, "explore", source.exploreName);
    // The resolver query's metrics view is not part of the url, so the explore's spec is not needed.
    const search = mapMetricsResolverQueryToUrlSearch({
      query: mapAIResolverPropsToMetricsResolverQuery(source.props, ""),
    }).toString();
    return search ? `${path}?${search}` : path;
  }

  const { request, options } = source;
  const { exploreName, metricsViewName, searchParams } =
    await getSearchParamsFromQuery(client, request, options);

  // backwards compatibility for older alerts created on metrics explore directly
  const path = getDashboardPath(ctx, "explore", exploreName || metricsViewName);
  const search = searchParams.toString();
  return search ? `${path}?${search}` : path;
}

function canMapExploreUrl(source: MappedExploreUrlSource) {
  if (source.type === "ai") return !!source.exploreName;
  return !!source.request.queryName && !!source.request.queryArgsJson;
}

async function getSearchParamsFromQuery(
  client: RuntimeClient,
  { exploreName, queryName, queryArgsJson, executionTime }: MapQueryRequest,
  {
    exploreProtoState,
    ignoreFilters = false,
    forceOpenPivot = false,
  }: MapQueryStateOptions,
) {
  if (!queryName || !queryArgsJson) {
    throw new Error("Required parameters are missing.");
  }

  const queryRequestProperties: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson),
  );

  switch (queryName) {
    case "MetricsViewAggregation": {
      // TODO: validate the request's shape instead of casting.
      const req = queryRequestProperties as V1MetricsViewAggregationRequest;
      return {
        exploreName,
        metricsViewName: req.metricsView ?? "",
        searchParams: await getUrlFromAggregationRequest({
          client,
          req,
          urlState: exploreProtoState,
          executionTime,
          ignoreFilters,
          forceOpenPivot,
        }),
      };
    }

    case "MetricsViewComparison": {
      const req = queryRequestProperties as V1MetricsViewComparisonRequest;
      return {
        exploreName,
        metricsViewName: req.metricsViewName ?? "",
        searchParams: await getUrlFromComparisonRequest({
          client,
          req,
          executionTime,
        }),
      };
    }

    // TODO
    // case "MetricsViewToplist":
    // case "MetricsViewRows":
    // case "MetricsViewTimeSeries":
  }

  throw new Error(
    "Failed to find metrics view name. Please check the format of the report.",
  );
}
