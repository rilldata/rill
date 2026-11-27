import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type {
  MapQueryRequest,
  MapQueryStateOptions,
  QueryRequests,
} from "@rilldata/web-common/features/explore-mappers/url-mappers/types.ts";
import {
  convertRequestKeysToCamelCase,
  type DashboardPathContext,
  getDashboardPath,
} from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { convertAggregationRequestToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-aggregation-request-to-url-params.ts";
import { convertComparisonRequestToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-comparison-request-to-url-params.ts";
import { convertMetricsResolverQueryToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-metrics-resolver-query-to-url-params.ts";
import {
  type AIResolverProps,
  mapAIResolverPropsToMetricsResolverQuery,
} from "@rilldata/web-common/features/explore-mappers/map-ai-resolver-props-to-metrics-resolver-query.ts";
import {
  getRuntimeServiceGetExploreQueryOptions,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

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
 * The url is validated when the explore is built from it,
 * so specs are only fetched where the mapping itself needs them (the time dimension),
 * or where the explore would silently drop a param (a time dimension it does not expose).
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
    // Relative time ranges are resolved to absolute ones against the execution time,
    // so a cached url stays valid; without one they are resolved against now, which moves on.
    staleTime:
      source.type === "query" && !source.request.executionTime ? 0 : Infinity,
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
    // The resolver query's metrics view is not part of the url,
    // but a filter on the time dimension is moved to the explore's time range.
    const exploreResp = await queryClient.fetchQuery(
      getRuntimeServiceGetExploreQueryOptions(client, {
        name: source.exploreName,
      }),
    );
    const search = convertMetricsResolverQueryToUrlParams(
      { query: mapAIResolverPropsToMetricsResolverQuery(source.props, "") },
      exploreResp.metricsView?.metricsView?.state?.validSpec?.timeDimension,
    ).toString();
    return search ? `${path}?${search}` : path;
  }

  const { request, options } = source;
  const { exploreName, metricsViewName, searchParams } =
    await convertQueryRequestToUrlParams(client, request, options);
  // backwards compatibility for older alerts created on metrics explore directly
  const targetExploreName = exploreName || metricsViewName;

  // The explore drops a time dimension it does not expose and falls back to the default one,
  // which would show a different time window than the query's.
  const timeDimension = searchParams.get(ExploreStateURLParams.TimeDimension);
  if (timeDimension) {
    const exploreResp = await queryClient.fetchQuery(
      getRuntimeServiceGetExploreQueryOptions(client, {
        name: targetExploreName,
      }),
    );
    if (
      !exploreResp.explore?.explore?.state?.validSpec?.dimensions?.includes(
        timeDimension,
      )
    ) {
      throw new Error(
        `The explore cannot show this query's time range on "${timeDimension}".`,
      );
    }
  }

  const path = getDashboardPath(ctx, "explore", targetExploreName);
  const search = searchParams.toString();
  return search ? `${path}?${search}` : path;
}

function canMapExploreUrl(source: MappedExploreUrlSource) {
  if (source.type === "ai") return !!source.exploreName;
  return !!source.request.queryName && !!source.request.queryArgsJson;
}

async function convertQueryRequestToUrlParams(
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
        searchParams: await convertAggregationRequestToUrlParams({
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
        searchParams: await convertComparisonRequestToUrlParams({
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
