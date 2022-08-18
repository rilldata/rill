/**
 * TODO: Instead of writing this file by hand, a better approach would be to use an OpenAPI spec and
 * autogenerate `svelte-query`-specific client code. One such tool is: https://orval.dev/guides/svelte-query
 */

import type {
  RuntimeTimeSeriesRequest,
  RuntimeTimeSeriesResponse,
} from "$common/rill-developer-service/MetricViewActions";
import { config } from "$lib/application-state-stores/application-store";

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
  request: RuntimeTimeSeriesRequest
): Promise<RuntimeTimeSeriesResponse> => {
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

// TODO...

// POST /api/v1/metric-views/{view-name}/totals

// TODO...
