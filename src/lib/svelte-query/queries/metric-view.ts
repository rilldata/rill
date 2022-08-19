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
} from "$common/rill-developer-service/MetricViewActions";
import { config } from "$lib/application-state-stores/application-store";
import type { QueryClient } from "@sveltestack/svelte-query";

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

export const getMetricViewMetaQueryKey = (metricViewId: string) => {
  return [`v1/metric-view/meta`, metricViewId];
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

export const getMetricViewTimeSeriesQueryKey = (metricViewId: string) => {
  return [`v1/metric-view/timeseries`, metricViewId];
};

// POST /api/v1/metric-views/{view-name}/toplist/{dimension}

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
  metricsDefId: string,
  dimensionId: string
) => {
  return [TopListId, metricsDefId, dimensionId];
};

export const invalidateMetricViewTopList = (
  queryClient: QueryClient,
  metricsDefId: string
) => {
  return queryClient.invalidateQueries([TopListId, metricsDefId]);
};

// POST /api/v1/metric-views/{view-name}/totals

// TODO...
