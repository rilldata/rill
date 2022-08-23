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
import type { QueryClient } from "@sveltestack/svelte-query";
import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

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

export const getMetricViewTimeSeriesRequest = (
  metricsExplorer: MetricsExplorerEntity,
  measures: Array<MeasureDefinitionEntity>
) => {
  const selectedMeasureIdSet = new Set(metricsExplorer.selectedMeasureIds);
  return {
    measures: measures
      .map((measure) => measure.id)
      .filter((measureId) => selectedMeasureIdSet.has(measureId)),
    filter: metricsExplorer.filters,
    time: {
      start: metricsExplorer?.selectedTimeRange?.start,
      end: metricsExplorer?.selectedTimeRange?.end,
      granularity: metricsExplorer?.selectedTimeRange?.interval,
    },
  };
};

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
  metricsDefId: string,
  dimensionId: string
) => {
  return [TopListId, metricsDefId, dimensionId];
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
  metricsDefId: string,
  isReferenceValue = false
) => {
  return [TotalsId, metricsDefId, isReferenceValue];
};

export const invalidateMetricViewData = (
  queryClient: QueryClient,
  metricsDefId: string
) => {
  queryClient.invalidateQueries([TopListId, metricsDefId]);
  queryClient.invalidateQueries([TotalsId, metricsDefId]);
  queryClient.invalidateQueries([TimeSeriesId, metricsDefId]);
};
