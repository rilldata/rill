import type { QueryObserverResult } from "@rilldata/svelte-query";
import { type CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type RpcStatus,
  type V1CanvasSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

export class CanvasResolvedSpec {
  canvasSpec: Readable<V1CanvasSpec | undefined>;
  isLoading: Readable<boolean>;
  metricViewNames: Readable<string[]>;

  getMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2[]>;

  getSimpleMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2[]>;

  getMeasureForMetricView: (
    measureName: string,
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2 | undefined>;

  getDimensionsForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2[]>;

  getDimensionForMetricView: (
    dimensionName: string,
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2 | undefined>;

  constructor(
    validSpecStore: Readable<QueryObserverResult<CanvasResponse, RpcStatus>>,
  ) {
    this.canvasSpec = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.data?.canvas;
    });

    this.isLoading = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.isLoading;
    });

    this.metricViewNames = derived(validSpecStore, ($validSpecStore) =>
      Object.keys($validSpecStore?.data?.metricsViews || {}),
    );

    this.getMeasuresForMetricView = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        if (!$validSpecStore.data) return [];
        const metricsViewData =
          $validSpecStore.data?.metricsViews[metricViewName];
        return metricsViewData?.state?.validSpec?.measures ?? [];
      });

    this.getSimpleMeasuresForMetricView = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        if (!$validSpecStore.data) return [];
        const metricsViewData =
          $validSpecStore.data?.metricsViews[metricViewName];

        return (
          metricsViewData?.state?.validSpec?.measures?.filter(
            (m) =>
              !m.window &&
              m.type !==
                MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
          ) ?? []
        );
      });

    this.getDimensionsForMetricView = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        if (!$validSpecStore.data) return [];
        const metricsViewData =
          $validSpecStore.data?.metricsViews[metricViewName];
        return metricsViewData?.state?.validSpec?.dimensions ?? [];
      });

    this.getMeasureForMetricView = (
      measureName: string,
      metricViewName: string,
    ) =>
      derived(this.getMeasuresForMetricView(metricViewName), (measures) => {
        return measures?.find((measure) => measure.name === measureName);
      });

    this.getDimensionForMetricView = (
      dimensionName: string,
      metricViewName: string,
    ) =>
      derived(this.getDimensionsForMetricView(metricViewName), (dimensions) => {
        return dimensions?.find(
          (d) => d.name === dimensionName || d.column === dimensionName,
        );
      });

    // export const useAllDimensionFromMetrics = (
    //   instanceId: string,
    //   metricsViewNames: string[],
    // ) => {
    //   const dimensionsByViewStores = metricsViewNames.map((viewName) =>
    //     useAllDimensionFromMetric(instanceId, viewName),
    //   );
    //   return derived(dimensionsByViewStores, ($dimensionsByViewStores) =>
    //     $dimensionsByViewStores
    //       .filter((dimensions) => dimensions?.data)
    //       .map((dimensions) => dimensions.data)
    //       .flat(),
    //   );
    // };
  }
}
