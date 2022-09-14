import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  MetricsViewMetaResponse,
  MetricsViewTimeSeriesRequest,
  MetricsViewTopListRequest,
  MetricsViewTotalsRequest,
} from "$common/rill-developer-service/MetricsViewActions";
import type { MetricsExplorerEntity } from "$lib/application-state-stores/explorer-stores";
import {
  selectMappedFilterFromMeta,
  selectMeasureNamesFromMeta,
} from "$lib/svelte-query/selectors/metrics-view";

export function getTimeSeriesRequest(
  meta: MetricsViewMetaResponse,
  metricsExplorer: MetricsExplorerEntity
): MetricsViewTimeSeriesRequest {
  return {
    measures: selectMeasureNamesFromMeta(
      meta,
      metricsExplorer.selectedMeasureIds
    ),
    filter: selectMappedFilterFromMeta(meta, metricsExplorer.filters),
    time: {
      start: metricsExplorer.selectedTimeRange?.start,
      end: metricsExplorer.selectedTimeRange?.end,
      granularity: metricsExplorer.selectedTimeRange?.interval,
    },
  };
}

export function getTopListRequest(
  meta: MetricsViewMetaResponse,
  measure: MeasureDefinitionEntity,
  metricsExplorer: MetricsExplorerEntity
): MetricsViewTopListRequest {
  return {
    measures: [measure.sqlName],
    limit: 15,
    offset: 0,
    sort: [
      {
        name: measure.sqlName,
        direction: "desc",
      },
    ],
    time: {
      start: metricsExplorer.selectedTimeRange?.start,
      end: metricsExplorer.selectedTimeRange?.end,
    },
    filter: selectMappedFilterFromMeta(meta, metricsExplorer.filters),
  };
}

export function getTotalsRequest(
  meta: MetricsViewMetaResponse,
  metricsExplorer: MetricsExplorerEntity,
  includeFilters = true
): MetricsViewTotalsRequest {
  return {
    measures: selectMeasureNamesFromMeta(
      meta,
      metricsExplorer.selectedMeasureIds
    ),
    ...(includeFilters
      ? { filter: selectMappedFilterFromMeta(meta, metricsExplorer.filters) }
      : {}),
    time: {
      start: metricsExplorer.selectedTimeRange?.start,
      end: metricsExplorer.selectedTimeRange?.end,
    },
  };
}
