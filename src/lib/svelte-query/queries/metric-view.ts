import type {
  RuntimeMetricsMetaResponse,
  RuntimeTimeSeriesRequest,
  RuntimeTimeSeriesResponse,
} from "$common/rill-developer-service/MetricViewActions";
import { config } from "$lib/application-state-stores/application-store";
import { fetchWrapper } from "$lib/util/fetchWrapper";

// GET /api/v1/metric-views/{view-name}/meta

export const getMetricViewMetadata = async (
  metricViewId: string
): Promise<RuntimeMetricsMetaResponse> => {
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

export const postMetricViewTimeSeries = (
  metricViewId: string,
  request: RuntimeTimeSeriesRequest
): Promise<RuntimeTimeSeriesResponse> => {
  return fetchWrapper(`v1/metric-views/${metricViewId}/timeseries`, "POST", {
    ...request,
  });
};

export const getMetricViewTimeSeriesQueryKey = (metricViewId: string) => {
  return [`v1/metric-views/${metricViewId}/timeseries`];
};

// POST /api/v1/metric-views/{view-name}/toplist/{dimension}

// TODO...

// POST /api/v1/metric-views/{view-name}/totals

// TODO...
