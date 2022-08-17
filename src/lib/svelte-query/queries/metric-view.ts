import type {
  RuntimeMetricsMetaResponse,
  RuntimeTimeSeriesRequest,
  RuntimeTimeSeriesResponse,
} from "$common/rill-developer-service/MetricViewActions";
import { fetchWrapper } from "$lib/util/fetchWrapper";

// GET /api/v1/metric-views/{view-name}/meta

export const getMetricViewMetadata = (
  metricViewId: string
): Promise<RuntimeMetricsMetaResponse> => {
  return fetchWrapper(`v1/metric-views/${metricViewId}/meta`, "GET");
};

export const getMetricViewMetaQueryKey = (metricViewId: string) => {
  return [`v1/metric-views/${metricViewId}/meta`];
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
