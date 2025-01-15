import type { QueryObserverResult } from "@rilldata/svelte-query";
import { type CanvasValidResponse } from "@rilldata/web-common/features/canvas/selector";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecDimensionV2,
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
  ) => Readable<MetricsViewSpecDimensionV2[]>;

  getSimpleMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2[]>;

  getDimensionsForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2[]>;

  constructor(
    validSpecStore: Readable<
      QueryObserverResult<CanvasValidResponse, RpcStatus>
    >,
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
  }
}
