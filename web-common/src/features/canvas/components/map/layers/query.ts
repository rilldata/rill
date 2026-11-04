import { canQueryWithTimeRange } from "@rilldata/web-common/features/components/charts/query-util";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import type { MapComponent, MapSpec } from "..";

export interface MapLayerFields {
  dimensions: V1MetricsViewAggregationDimension[];
  measures: V1MetricsViewAggregationMeasure[];
}

/**
 * Creates the aggregation query that feeds one layer of a map.
 * `getFields` returns the layer's dimensions and measures,
 * or null while the layer's part of the spec is incomplete, which disables the query.
 */
export function createMapLayerQuery(
  runtimeClient: RuntimeClient,
  component: MapComponent,
  getFields: (spec: Partial<MapSpec>) => MapLayerFields | null,
) {
  const { specStore, timeAndFilterStore, dataEnabled } = component;

  const queryOptionsStore = derived(
    [specStore, timeAndFilterStore, dataEnabled],
    ([specVal, $timeAndFilterStore, $visible]) => {
      const spec = specVal ?? ({} as Partial<MapSpec>);
      const fields = getFields(spec);
      const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;

      const enabled =
        $visible &&
        !!spec.metrics_view &&
        !!fields &&
        canQueryWithTimeRange(hasTimeSeries, timeRange);

      return getQueryServiceMetricsViewAggregationQueryOptions(
        runtimeClient,
        {
          metricsView: spec.metrics_view ?? "",
          dimensions: fields?.dimensions ?? [],
          measures: fields?.measures ?? [],
          limit: "5000",
          where,
          timeRange: hasTimeSeries ? timeRange : undefined,
        },
        {
          query: {
            enabled,
            placeholderData: keepPreviousData,
          },
        },
      );
    },
  );

  return createQuery(queryOptionsStore);
}
