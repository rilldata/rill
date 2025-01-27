import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1CanvasSpec,
  type V1ComponentSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, get, type Readable } from "svelte/store";

export class CanvasResolvedSpec {
  canvasSpec: Readable<V1CanvasSpec | undefined>;
  isLoading: Readable<boolean>;
  metricViewNames: Readable<string[]>;

  getMetricsViewFromName: (
    metricViewName: string,
  ) => Readable<V1MetricsViewSpec | undefined>;
  /** Measure Selectors */
  getMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2[]>;

  getSimpleMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2[]>;

  getMeasureForMetricView: (
    measureName: string | undefined,
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasureV2 | undefined>;

  allSimpleMeasures: Readable<MetricsViewSpecMeasureV2[]>;

  metricsViewMeasureMap: Readable<Record<string, Set<string>>>;

  /** Dimension Selectors */
  getDimensionsForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2[]>;

  getDimensionForMetricView: (
    dimensionName: string,
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimensionV2 | undefined>;

  getTimeDimensionForMetricView: (
    metricViewName: string,
  ) => Readable<string | undefined>;

  allDimensions: Readable<MetricsViewSpecDimensionV2[]>;
  metricsViewDimensionsMap: Readable<Record<string, Set<string>>>;

  /** Component Selectors */
  getComponentResource: (
    componentName: string,
  ) => Readable<V1ComponentSpec | undefined>;

  constructor(validSpecStore: CanvasSpecResponseStore) {
    this.canvasSpec = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.data?.canvas;
    });

    this.isLoading = derived(validSpecStore, ($validSpecStore) => {
      return $validSpecStore.isLoading;
    });

    this.metricViewNames = derived(validSpecStore, ($validSpecStore) =>
      Object.keys($validSpecStore?.data?.metricsViews || {}),
    );

    this.getMetricsViewFromName = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        const metricsView = $validSpecStore.data?.metricsViews[metricViewName];
        if (!metricsView) return;
        return metricsView.state?.validSpec;
      });

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

        return this.filterSimpleMeasures(
          metricsViewData?.state?.validSpec?.measures,
        );
      });

    this.allSimpleMeasures = derived(validSpecStore, ($validSpecStore) => {
      if (!$validSpecStore.data) return [];
      const measures = Object.values(
        $validSpecStore.data.metricsViews || {},
      ).flatMap((metricsView) =>
        this.filterSimpleMeasures(metricsView?.state?.validSpec?.measures),
      );
      const uniqueByName = new Map<string, MetricsViewSpecMeasureV2>();
      for (const measure of measures) {
        uniqueByName.set(measure.name as string, measure);
      }
      return [...uniqueByName.values()];
    });

    this.allDimensions = derived(validSpecStore, ($validSpecStore) => {
      if (!$validSpecStore.data) return [];
      const dimensions = Object.values(
        $validSpecStore.data.metricsViews || {},
      ).flatMap(
        (metricsView) => metricsView?.state?.validSpec?.dimensions || [],
      );
      const uniqueByName = new Map<string, MetricsViewSpecDimensionV2>();
      for (const dimension of dimensions) {
        uniqueByName.set(
          (dimension.name || dimension.column) as string,
          dimension,
        );
      }
      return [...uniqueByName.values()];
    });

    this.getDimensionsForMetricView = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        if (!$validSpecStore.data) return [];
        const metricsViewData =
          $validSpecStore.data?.metricsViews[metricViewName];
        return metricsViewData?.state?.validSpec?.dimensions ?? [];
      });

    this.getMeasureForMetricView = (
      measureName: string | undefined,
      metricViewName: string,
    ) =>
      derived(this.getMeasuresForMetricView(metricViewName), (measures) => {
        if (!measureName) return;
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

    this.getTimeDimensionForMetricView = (metricViewName: string) =>
      derived(validSpecStore, ($validSpecStore) => {
        if (!$validSpecStore.data) return undefined;
        const metricsViewData =
          $validSpecStore.data?.metricsViews[metricViewName];
        return metricsViewData?.state?.validSpec?.timeDimension;
      });

    this.metricsViewMeasureMap = derived(validSpecStore, ($validSpecStore) => {
      const metricsViewMeasureMap: Record<string, Set<string>> = {};
      for (const [metricViewName, metricsViewData] of Object.entries(
        $validSpecStore.data?.metricsViews || {},
      )) {
        metricsViewMeasureMap[metricViewName] = new Set(
          metricsViewData?.state?.validSpec?.measures?.map(
            (m) => m.name as string,
          ) || [],
        );
      }
      return metricsViewMeasureMap;
    });

    this.metricsViewDimensionsMap = derived(
      validSpecStore,
      ($validSpecStore) => {
        const metricsViewDimensionMap: Record<string, Set<string>> = {};
        for (const [metricViewName, metricsViewData] of Object.entries(
          $validSpecStore.data?.metricsViews || {},
        )) {
          metricsViewDimensionMap[metricViewName] = new Set(
            metricsViewData?.state?.validSpec?.dimensions?.map(
              (d) => (d.name || d.column) as string,
            ) || [],
          );
        }
        return metricsViewDimensionMap;
      },
    );

    this.getComponentResource = (componentName: string) => {
      return derived(validSpecStore, ($validSpecStore) => {
        return $validSpecStore?.data?.components?.[componentName]?.component
          ?.spec;
      });
    };
  }

  getDimensionsFromMeasure(measureName: string): MetricsViewSpecDimensionV2[] {
    const metricsMeasureMap = get(this.metricsViewMeasureMap);
    let metricViewName: string | undefined;
    for (const [key, value] of Object.entries(metricsMeasureMap)) {
      if (value.has(measureName)) metricViewName = key;
    }

    if (metricViewName)
      return get(this.getDimensionsForMetricView(metricViewName));
    return [];
  }

  getMetricsViewNamesForDimension(dimensionName: string): string[] {
    const metricsDimensionMap = get(this.metricsViewDimensionsMap);
    const metricViewNames: string[] = [];
    for (const [key, value] of Object.entries(metricsDimensionMap)) {
      if (value.has(dimensionName)) metricViewNames.push(key);
    }
    return metricViewNames;
  }

  private filterSimpleMeasures = (
    measures: MetricsViewSpecMeasureV2[] | undefined,
  ): MetricsViewSpecMeasureV2[] => {
    return (
      measures?.filter(
        (m) =>
          !m.window &&
          m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
      ) || []
    );
  };
}
