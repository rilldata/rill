/**
 * TODO: Instead of writing this file by hand, a better approach would be to use an OpenAPI spec and
 * autogenerate `svelte-query`-specific client code. One such tool is: https://orval.dev/guides/svelte-query
 */

import type {
  MetricsViewMetaResponse,
  MetricsViewTimeSeriesRequest,
  MetricsViewTimeSeriesResponse,
  MetricsViewTopListRequest,
  MetricsViewTopListResponse,
  MetricsViewTotalsRequest,
  MetricsViewTotalsResponse,
} from "$common/rill-developer-service/MetricsViewActions";
import { config } from "$lib/application-state-stores/application-store";
import { QueryClient, useQuery } from "@sveltestack/svelte-query";

async function fetchUrl(path: string, method: string, body?) {
  const resp = await fetch(`${config.server.serverUrl}/api/v1/${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json.messages[0].message);
    return Promise.reject(err);
  }
  return json;
}

// GET /api/v1/metrics-views/{view-name}/meta

export const getMetricsViewMetadata = async (
  metricViewId: string
): Promise<MetricsViewMetaResponse> => {
  const json = await fetchUrl(`metrics-views/${metricViewId}/meta`, "GET");
  json.id = metricViewId;
  return json;
};

const MetaId = `v1/metrics-view/meta`;
export const getMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useMetaQuery = (metricViewId: string) => {
  const metaQueryKey = getMetaQueryKey(metricViewId);
  const metaQueryFn = () => getMetricsViewMetadata(metricViewId);
  const metaQueryOptions = {
    enabled: !!metricViewId,
  };
  return useQuery<MetricsViewMetaResponse, Error>(
    metaQueryKey,
    metaQueryFn,
    metaQueryOptions
  );
};

export const selectDimensionFromMeta = (
  meta: MetricsViewMetaResponse,
  dimensionId: string
) => meta?.dimensions?.find((dimension) => dimension.id === dimensionId);
export const selectMeasureFromMeta = (
  meta: MetricsViewMetaResponse,
  measureId: string
) => meta?.measures?.find((measure) => measure.id === measureId);

// POST /api/v1/metrics-views/{view-name}/timeseries

export const getMetricsViewTimeSeries = async (
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
): Promise<MetricsViewTimeSeriesResponse> => {
  return fetchUrl(`metrics-views/${metricViewId}/timeseries`, "POST", request);
};

const TimeSeriesId = `v1/metrics-view/timeseries`;
export const getTimeSeriesQueryKey = (
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
) => {
  return [TimeSeriesId, metricViewId, request];
};

export const useTimeSeriesQuery = (
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
) => {
  const timeSeriesQueryKey = getTimeSeriesQueryKey(metricViewId, request);
  const timeSeriesQueryFn = () =>
    getMetricsViewTimeSeries(metricViewId, request);
  const timeSeriesQueryOptions = {
    staleTime: 1000 * 30,
    enabled: !!(metricViewId && request.measures && request.time),
  };
  return useQuery<MetricsViewTimeSeriesResponse, Error>(
    timeSeriesQueryKey,
    timeSeriesQueryFn,
    timeSeriesQueryOptions
  );
};

// POST /api/v1/metrics-views/{view-name}/toplist/{dimension}

export const getMetricsViewTopList = async (
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
    `metrics-views/${metricViewId}/toplist/${dimensionId}`,
    "POST",
    request
  );
};

const TopListId = `v1/metrics-view/toplist`;
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
    return getMetricsViewTopList(metricsDefId, dimensionId, requestParameter);
  };
  const topListQueryOptions = getTopListQueryOptions(
    metricsDefId,
    dimensionId,
    requestParameter
  );
  return useQuery(topListQueryKey, topListQueryFn, topListQueryOptions);
}

// POST /api/v1/metrics-views/{view-name}/totals

export const getMetricsViewTotals = async (
  metricViewId: string,
  request: MetricsViewTotalsRequest
): Promise<MetricsViewTotalsResponse> => {
  return fetchUrl(`metrics-views/${metricViewId}/totals`, "POST", request);
};

const TotalsId = `v1/metrics-view/totals`;
export const getTotalsQueryKey = (
  metricViewId: string,
  request: MetricsViewTotalsRequest
) => {
  return [TotalsId, metricViewId, request];
};

export const useTotalsQuery = (
  metricViewId: string,
  request: MetricsViewTotalsRequest
) => {
  const totalsQueryKey = getTotalsQueryKey(metricViewId, request);
  const totalsQueryFn = () => getMetricsViewTotals(metricViewId, request);

  return useQuery<MetricsViewTotalsResponse, Error>(
    totalsQueryKey,
    totalsQueryFn,
    {
      staleTime: 30 * 1000,
      enabled: !!(metricViewId && request.measures && request.time),
    }
  );
};

// invalidation helpers

export const invalidateMetricsView = async (
  queryClient: QueryClient,
  metricViewId: string
) => {
  // wait for meta to be invalidated
  await queryClient.refetchQueries([MetaId, metricViewId]);
  // invalidateMetricsViewData(queryClient, metricViewId);
};

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  metricViewId: string
) => {
  queryClient.refetchQueries([TopListId, metricViewId]);
  queryClient.refetchQueries([TotalsId, metricViewId]);
  queryClient.refetchQueries([TimeSeriesId, metricViewId]);
};
