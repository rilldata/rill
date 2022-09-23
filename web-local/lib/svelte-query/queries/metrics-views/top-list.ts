import type { RootConfig } from "../../../../common/config/RootConfig";
import type {
  MetricsViewTopListRequest,
  MetricsViewTopListResponse,
} from "../../../../common/rill-developer-service/MetricsViewActions";
import { fetchUrl } from "../fetch-url";
import { useQuery } from "@sveltestack/svelte-query";

// POST /api/v1/metrics-views/{view-name}/toplist/{dimension}

export const getMetricsViewTopList = async (
  config: RootConfig,
  metricViewId: string,
  dimensionId: string,
  request: MetricsViewTopListRequest
): Promise<MetricsViewTopListResponse> => {
  if (!metricViewId || !dimensionId)
    return {
      meta: [],
      data: [],
    };

  return fetchUrl(
    config.server.exploreUrl,
    `metrics-views/${metricViewId}/toplist/${dimensionId}`,
    "POST",
    request
  );
};

export const TopListId = `v1/metrics-view/toplist`;

export const getTopListQueryKey = (
  metricViewId: string,
  dimensionId: string,
  request: MetricsViewTopListRequest
) => {
  return [TopListId, metricViewId, dimensionId, request];
};

function getTopListQueryOptions(
  metricsDefId,
  dimensionId,
  topListQueryRequest
) {
  return {
    staleTime: 30 * 1000,
    enabled: !!(
      metricsDefId &&
      dimensionId &&
      topListQueryRequest.limit &&
      topListQueryRequest.measures.length >= 1 &&
      topListQueryRequest.offset !== undefined &&
      topListQueryRequest.sort &&
      topListQueryRequest.time
    ),
  };
}

/** custom hook to fetch a toplist result set, given a metricsDefId,
 * dimensionId,
 * and a request parameter.
 * The request parameter matches the API signature needed for the toplist request.
 */
export function useTopListQuery(
  config: RootConfig,
  metricsDefId: string,
  dimensionId: string,
  requestParameter: MetricsViewTopListRequest
) {
  const topListQueryKey = getTopListQueryKey(
    metricsDefId,
    dimensionId,
    requestParameter
  );
  const topListQueryFn = () => {
    return getMetricsViewTopList(
      config,
      metricsDefId,
      dimensionId,
      requestParameter
    );
  };
  const topListQueryOptions = getTopListQueryOptions(
    metricsDefId,
    dimensionId,
    requestParameter
  );
  return useQuery(topListQueryKey, topListQueryFn, topListQueryOptions);
}
