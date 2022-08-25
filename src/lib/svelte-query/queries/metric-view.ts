/**
 * TODO: Instead of writing this file by hand, a better approach would be to use an OpenAPI spec and
 * autogenerate `svelte-query`-specific client code. One such tool is: https://orval.dev/guides/svelte-query
 */

import type {
  MetricViewMetaResponse,
  MetricViewTimeSeriesRequest,
  MetricViewTimeSeriesResponse,
  MetricViewTopListRequest,
  MetricViewTopListResponse,
  MetricViewTotalsRequest,
  MetricViewTotalsResponse,
} from "$common/rill-developer-service/MetricViewActions";
import { config } from "$lib/application-state-stores/application-store";
import type { MetricsExplorerEntity } from "$lib/application-state-stores/explorer-stores";
import {
  QueryClient,
  useQuery,
  UseQueryOptions,
} from "@sveltestack/svelte-query";

// GET /api/v1/metric-views/{view-name}/meta

export const getMetricViewMetadata = async (
  metricViewId: string
): Promise<MetricViewMetaResponse> => {
  const resp = await fetch(
    `${config.server.serverUrl}/api/v1/metric-views/${metricViewId}/meta`,
    {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    }
  );
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json);
    return Promise.reject(err);
  }
  return json;
};

const MetaId = `v1/metric-view/meta`;
export const getMetricViewMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useGetMetricViewMeta = (
  metricViewId: string,
  queryOptions?: UseQueryOptions<MetricViewMetaResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ?? getMetricViewMetaQueryKey(metricViewId);
  const queryFn = () => getMetricViewMetadata(metricViewId);
  const query = useQuery<MetricViewMetaResponse, Error>(
    queryKey,
    queryFn,
    queryOptions
  );
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/timeseries

export const getMetricViewTimeSeries = async (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest
): Promise<MetricViewTimeSeriesResponse> => {
  const resp = await fetch(
    `${config.server.serverUrl}/api/v1/metric-views/${metricViewId}/timeseries`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    }
  );
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json);
    return Promise.reject(err);
  }
  return json;
};

const TimeSeriesId = `v1/metric-view/timeseries`;
export const getMetricViewTimeSeriesQueryKey = (metricViewId: string) => {
  return [TimeSeriesId, metricViewId];
};

export const useGetMetricViewTimeSeries = (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest,
  queryOptions?: UseQueryOptions<MetricViewTimeSeriesResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ?? getMetricViewTimeSeriesQueryKey(metricViewId);
  const queryFn = () => getMetricViewTimeSeries(metricViewId, request);
  const query = useQuery<MetricViewTimeSeriesResponse, Error>(
    queryKey,
    queryFn,
    queryOptions
  );
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/toplist/{dimension}
export const getTopListRequest = (
  metricsExplorer: MetricsExplorerEntity
): MetricViewTopListRequest => {
  return {
    measures: [metricsExplorer.leaderboardMeasureId],
    limit: 10,
    offset: 0,
    sort: [],
    time: {
      start: metricsExplorer.selectedTimeRange?.start,
      end: metricsExplorer.selectedTimeRange?.end,
    },
    filter: metricsExplorer.filters,
  };
};

export const getMetricViewTopList = async (
  metricViewId: string,
  dimensionId: string,
  request: MetricViewTopListRequest
): Promise<MetricViewTopListResponse> => {
  if (!metricViewId || !dimensionId)
    return {
      meta: [],
      data: [],
    };

  const resp = await fetch(
    `${config.server.serverUrl}/api/v1/metric-views/${metricViewId}/toplist/${dimensionId}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    }
  );
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json);
    return Promise.reject(err);
  }
  return json;
};

const TopListId = `v1/metric-view/toplist`;
export const getMetricViewTopListQueryKey = (
  metricViewId: string,
  dimensionId: string
) => {
  return [TopListId, metricViewId, dimensionId];
};

// POST /api/v1/metric-views/{view-name}/totals
export const getTotalsRequest = (
  metricsExplorer: MetricsExplorerEntity,
  noFilters = false
): MetricViewTotalsRequest => {
  return {
    measures: metricsExplorer.selectedMeasureIds,
    filter: noFilters ? undefined : metricsExplorer.filters,
    time: {
      start: metricsExplorer.selectedTimeRange?.start,
      end: metricsExplorer.selectedTimeRange?.end,
    },
  };
};

export const getMetricViewTotals = async (
  metricViewId: string,
  request: MetricViewTotalsRequest
): Promise<MetricViewTotalsResponse> => {
  const resp = await fetch(
    `${config.server.serverUrl}/api/v1/metric-views/${metricViewId}/totals`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    }
  );
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json);
    return Promise.reject(err);
  }
  return json;
};

const TotalsId = `v1/metric-view/totals`;
export const getMetricViewTotalsQueryKey = (
  metricViewId: string,
  isReferenceValue = false
) => {
  return [TotalsId, metricViewId, isReferenceValue];
};

export const invalidateMetricView = async (
  queryClient: QueryClient,
  metricViewId: string
) => {
  // wait for meta to be invalidated
  await queryClient.invalidateQueries([MetaId, metricViewId]);
  invalidateMetricViewData(queryClient, metricViewId);
};

export const invalidateMetricViewData = (
  queryClient: QueryClient,
  metricViewId: string
) => {
  queryClient.invalidateQueries([TopListId, metricViewId]);
  queryClient.invalidateQueries([TotalsId, metricViewId]);
  queryClient.invalidateQueries([TimeSeriesId, metricViewId]);
};
