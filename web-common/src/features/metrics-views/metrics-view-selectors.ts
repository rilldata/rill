import {
  MetricsViewSpecDimensionType,
  MetricsViewSpecMeasureType,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsView,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { derived, get, type Readable } from "svelte/store";
import {
  ResourceKind,
  useFilteredResources,
} from "../entity-management/resource-selectors";

type MetricsViewsData = Readable<Record<string, V1MetricsView | undefined>>;

export class MetricsViewSelectors {
  metricViewNames: Readable<string[]>;

  getMetricsViewFromName: (metricViewName: string) => Readable<{
    metricsView: V1MetricsViewSpec | undefined;
    isLoading: boolean;
  }>;

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

  allMetricsViews: ReturnType<
    typeof useFilteredResources<Array<V1Resource | undefined>>
  >;

  constructor(instanceId: string, metricsViewsData?: MetricsViewsData) {
    this.allMetricsViews = useFilteredResources(
      instanceId,
      ResourceKind.MetricsView,
    );

    // If metricsViewsData is not provided, create it from allMetricsViews
    const metricsViewResources =
      metricsViewsData ??
      derived(this.allMetricsViews, ($metricsViews) => {
        const metricsViewsMap: Record<string, V1MetricsView | undefined> = {};
        if ($metricsViews?.data) {
          for (const resource of $metricsViews.data) {
            const name = resource?.meta?.name?.name;
            if (name && resource?.metricsView) {
              metricsViewsMap[name] = resource.metricsView;
            }
          }
        }
        return metricsViewsMap;
      });

    this.metricViewNames = derived(
      metricsViewResources,
      ($metricsViewResources) => Object.keys($metricsViewResources || {}),
    );

    this.getMetricsViewFromName = (metricViewName: string) =>
      derived(this.allMetricsViews, ($metricsViews) => {
        return {
          metricsView: $metricsViews?.data?.find(
            (res) => res?.meta?.name?.name === metricViewName,
          )?.metricsView?.state?.validSpec,
          isLoading: $metricsViews?.isLoading,
        };
      });

    this.getMeasuresForMetricView = (metricViewName: string) =>
      derived(metricsViewResources, ($metricsViewResources) => {
        if (!$metricsViewResources) return [];
        const metricsView = $metricsViewResources[metricViewName];
        return metricsView?.state?.validSpec?.measures ?? [];
      });

    this.getSimpleMeasuresForMetricView = (metricViewName: string) =>
      derived(metricsViewResources, ($metricsViewResources) => {
        if (!$metricsViewResources) return [];
        const metricsView = $metricsViewResources[metricViewName];

        return this.filterSimpleMeasures(
          metricsView?.state?.validSpec?.measures,
        );
      });

    this.allSimpleMeasures = derived(
      metricsViewResources,
      ($metricsViewResources) => {
        if (!$metricsViewResources) return [];
        const measures = Object.values($metricsViewResources || {}).flatMap(
          (metricsView) =>
            this.filterSimpleMeasures(metricsView?.state?.validSpec?.measures),
        );
        const uniqueByName = new Map<string, MetricsViewSpecMeasure>();
        for (const measure of measures) {
          uniqueByName.set(measure.name as string, measure);
        }
        return [...uniqueByName.values()];
      },
    );

    this.allDimensions = derived(
      metricsViewResources,
      ($metricsViewResources) => {
        if (!$metricsViewResources) return [];
        const dimensions = Object.values($metricsViewResources || {})
          .flatMap(
            (metricsView) => metricsView?.state?.validSpec?.dimensions || [],
          )
          .filter(
            (d) => d.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
          );
        const uniqueByName = new Map<string, MetricsViewSpecDimension>();
        for (const dimension of dimensions) {
          uniqueByName.set(
            (dimension.name || dimension.column) as string,
            dimension,
          );
        }
        return [...uniqueByName.values()];
      },
    );

    this.getDimensionsForMetricView = (metricViewName: string) =>
      derived(metricsViewResources, ($metricsViewResources) => {
        if (!$metricsViewResources) return [];
        const metricsView = $metricsViewResources[metricViewName];
        return (
          metricsView?.state?.validSpec?.dimensions?.filter(
            (d) => d.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
          ) ?? []
        );
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
      derived(metricsViewResources, ($metricsViewResources) => {
        if (!$metricsViewResources) return undefined;
        const metricsView = $metricsViewResources[metricViewName];
        return metricsView?.state?.validSpec?.timeDimension;
      });

    this.metricsViewMeasureMap = derived(
      metricsViewResources,
      ($metricsViewResources) => {
        const metricsViewMeasureMap: Record<string, Set<string>> = {};
        for (const [metricViewName, metricsView] of Object.entries(
          $metricsViewResources || {},
        )) {
          metricsViewMeasureMap[metricViewName] = new Set(
            metricsView?.state?.validSpec?.measures?.map(
              (m) => m.name as string,
            ) || [],
          );
        }
        return metricsViewMeasureMap;
      },
    );

    this.metricsViewDimensionsMap = derived(
      metricsViewResources,
      ($metricsViewResources) => {
        const metricsViewDimensionMap: Record<string, Set<string>> = {};
        for (const [metricViewName, metricsView] of Object.entries(
          $metricsViewResources || {},
        )) {
          metricsViewDimensionMap[metricViewName] = new Set(
            metricsView?.state?.validSpec?.dimensions
              ?.filter(
                (d) =>
                  d.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
              )
              ?.map((d) => (d.name || d.column) as string) || [],
          );
        }
        return metricsViewDimensionMap;
      },
    );
  }

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
