import type { RootConfig } from "$web-local/common/config/RootConfig";
import type {
  MetricsViewTimeSeriesRequest,
  MetricsViewTimeSeriesResponse,
} from "$web-local/common/rill-developer-service/MetricsViewActions";
import { fetchUrl } from "../fetch-url";
import { useQuery } from "@sveltestack/svelte-query";

// POST /api/v1/metrics-views/{view-name}/timeseries

export const getMetricsViewTimeSeries = async (
  config: RootConfig,
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
): Promise<MetricsViewTimeSeriesResponse> => {
  return fetchUrl(
    config.server.exploreUrl,
    `metrics-views/${metricViewId}/timeseries`,
    "POST",
    request
  );
};

export const TimeSeriesId = `v1/metrics-view/timeseries`;

export const getTimeSeriesQueryKey = (
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
) => {
  return [TimeSeriesId, metricViewId, request];
};

export const useTimeSeriesQuery = (
  config: RootConfig,
  metricViewId: string,
  request: MetricsViewTimeSeriesRequest
) => {
  const timeSeriesQueryKey = getTimeSeriesQueryKey(metricViewId, request);
  const timeSeriesQueryFn = () =>
    getMetricsViewTimeSeries(config, metricViewId, request);
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
