import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1CanvasSpec,
  type V1ComponentSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, get, type Readable } from "svelte/store";

export class CanvasResolvedSpec {
  canvasSpec: Readable<V1CanvasSpec | undefined>;
  components: Readable<Record<string, V1ComponentSpec | undefined>>;
  isLoading: Readable<boolean>;
  metricViewNames: Readable<string[]>;

  getMetricsViewFromName: (
    metricViewName: string,
  ) => Readable<V1MetricsViewSpec | undefined>;
  /** Measure Selectors */
  getMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasure[]>;

  getSimpleMeasuresForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasure[]>;

  getMeasureForMetricView: (
    measureName: string | undefined,
    metricViewName: string,
  ) => Readable<MetricsViewSpecMeasure | undefined>;

  allSimpleMeasures: Readable<MetricsViewSpecMeasure[]>;

  metricsViewMeasureMap: Readable<Record<string, Set<string>>>;

  /** Dimension Selectors */
  getDimensionsForMetricView: (
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimension[]>;

  getDimensionForMetricView: (
    dimensionName: string,
    metricViewName: string,
  ) => Readable<MetricsViewSpecDimension | undefined>;

  getTimeDimensionForMetricView: (
    metricViewName: string,
  ) => Readable<string | undefined>;

  allDimensions: Readable<MetricsViewSpecDimension[]>;
  metricsViewDimensionsMap: Readable<Record<string, Set<string>>>;

  /** Component Selectors */
  getComponentResourceFromName: (
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

    this.components = derived(validSpecStore, ($validSpecStore) => {
      const componentResources = $validSpecStore.data?.components || {};
      const components: Record<string, V1ComponentSpec | undefined> = {};

      for (const [componentName, resource] of Object.entries(
        componentResources,
      )) {
        components[componentName] = resource.component?.state?.validSpec;
      }

      return components;
    });

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
      const uniqueByName = new Map<string, MetricsViewSpecMeasure>();
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
      const uniqueByName = new Map<string, MetricsViewSpecDimension>();
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

    this.getComponentResourceFromName = (componentName: string) => {
      return derived(this.components, (components) => {
        return components[componentName];
      });
    };
  }

  getComponentNameFromPos = (pos: {
    row: number;
    column: number;
  }): Readable<string | undefined> => {
    return derived(this.canvasSpec, (canvasSpec) => {
      const componentName =
        canvasSpec?.rows?.[pos.row]?.items?.[pos.column]?.component;
      if (!componentName) return undefined;
      return componentName;
    });
  };

  getComponentFromIndex = (pos: {
    row: number;
    column: number;
  }): Readable<V1ComponentSpec | undefined> => {
    const componentName = this.getComponentNameFromPos(pos);
    return derived(
      [componentName, this.components],
      ([componentName, components]) => {
        if (!componentName) return undefined;
        return components[componentName];
      },
    );
  };

  getDimensionsFromMeasure = (
    measureName: string,
  ): MetricsViewSpecDimension[] => {
    const metricsMeasureMap = get(this.metricsViewMeasureMap);
    let metricViewName: string | undefined;
    for (const [key, value] of Object.entries(metricsMeasureMap)) {
      if (value.has(measureName)) metricViewName = key;
    }

    if (metricViewName)
      return get(this.getDimensionsForMetricView(metricViewName));
    return [];
  };

  getMetricsViewNamesForDimension = (dimensionName: string): string[] => {
    const metricsDimensionMap = get(this.metricsViewDimensionsMap);
    const metricViewNames: string[] = [];
    for (const [key, value] of Object.entries(metricsDimensionMap)) {
      if (value.has(dimensionName)) metricViewNames.push(key);
    }
    return metricViewNames;
  };

  private filterSimpleMeasures = (
    measures: MetricsViewSpecMeasure[] | undefined,
  ): MetricsViewSpecMeasure[] => {
    return (
      measures?.filter(
        (m) =>
          !m.window &&
          m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
      ) || []
    );
  };
}
