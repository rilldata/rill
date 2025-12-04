import type {
  CanvasComponentType,
  ComponentSize,
  ComponentSpec,
} from "@rilldata/web-common/features/canvas/components/types";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type {
  V1Expression,
  V1Resource,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { ComponentType, SvelteComponent } from "svelte";
import { derived, get, writable, type Writable } from "svelte/store";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "../../dashboards/stores/filter-utils";
import type {
  ComparisonTimeRangeState,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import type {
  CanvasEntity,
  ComponentPath,
  SearchParamsStore,
} from "../stores/canvas-entity";
import { Filters } from "../stores/filters";
import { TimeControls } from "../stores/time-control";
import { ExploreStateURLParams } from "../../dashboards/url-state/url-params";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copoy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Widget specific dimension and measure filters
  localFilters: Filters;
  // Widget specific time filters
  localTimeControls: TimeControls;

  visible = writable(false);

  abstract type: CanvasComponentType;
  // Component responsible for DOM rendering
  abstract component: ComponentType<SvelteComponent>;
  // Will be deprecated
  abstract minSize: ComponentSize;
  // Will be deprecated
  abstract defaultSize: ComponentSize;
  // Parameters to reset when the metrics_view changes
  abstract resetParams: string[];
  // Minimum condition needed for the component to be rendered
  abstract isValid(spec: T): boolean;
  // Configuration for the sidebar editor
  abstract inputParams(type?: CanvasComponentType): InputParams<T>;

  getExploreTransformerProperties?(): Partial<ExploreState>;

  constructor(
    resource: V1Resource,
    public parent: CanvasEntity,
    path: ComponentPath,
    public defaultSpec: T,
  ) {
    const yamlSpec = resource.component?.state?.validSpec?.rendererProperties;

    const mergedSpec = { ...defaultSpec, ...yamlSpec };
    this.specStore = writable(mergedSpec);
    this.pathInYAML = path;

    this.resource.set(resource);
    this.id = resource.meta?.name?.name as string;

    const yamlTimeFilterStore: SearchParamsStore = (() => {
      const store = derived(this.specStore, (spec) => {
        return new URLSearchParams(spec?.["time_filters"] ?? "");
      });
      return {
        subscribe: store.subscribe,
        set: (key: string, value: string | undefined) => {
          const searchParams = get(store);

          if (value === undefined || value === null || value === "") {
            searchParams.delete(key);
          } else {
            searchParams.set(key, value);
          }

          this.updateProperty(
            "time_filters" as AllKeys<T>,
            searchParams.toString() as T[AllKeys<T>],
          );
          return true;
        },
        clearAll: () => {
          const searchParams = get(store);

          searchParams.forEach((_, key) => {
            searchParams.delete(key);
          });

          this.updateProperty(
            "time_filters" as AllKeys<T>,
            searchParams.toString() as T[AllKeys<T>],
          );
        },
      };
    })();

    const yamlDimensionFilterStore: SearchParamsStore = (() => {
      const store = derived(this.specStore, (spec) => {
        const dimensionFiltersString = spec?.["dimension_filters"] ?? "";
        const searchParams = new URLSearchParams();

        if (dimensionFiltersString) {
          searchParams.set(
            ExploreStateURLParams.Filters,
            dimensionFiltersString,
          );
        }

        return searchParams;
      });
      return {
        subscribe: store.subscribe,
        set: (key: string, value: string) => {
          const searchParams = get(store);

          searchParams.set(key, value);

          this.updateProperty(
            "dimension_filters" as AllKeys<T>,
            value as T[AllKeys<T>],
          );

          return true;
        },
        clearAll: () => {
          const searchParams = get(store);

          searchParams.forEach((_, key) => {
            searchParams.delete(key);
          });

          this.updateProperty(
            "dimension_filters" as AllKeys<T>,
            searchParams.toString() as T[AllKeys<T>],
          );
        },
      };
    })();

    this.localTimeControls = new TimeControls(
      this.parent.specStore,
      yamlTimeFilterStore,
      this.id,
      this.parent.name,
    );

    this.localFilters = new Filters(
      this.parent.metricsView,
      yamlDimensionFilterStore,
      this.id,
    );
  }

  update(resource: V1Resource, path: ComponentPath) {
    const yamlSpec = resource.component?.state?.validSpec
      ?.rendererProperties as T;
    this.resource.set(resource);
    this.pathInYAML = path;
    this.specStore.set(yamlSpec);
  }

  get timeAndFilterStore() {
    return derived(
      [
        this.parent.timeControls.timeRangeStateStore,
        this.localTimeControls.timeRangeStateStore,
        this.parent.timeControls.comparisonRangeStateStore,
        this.localTimeControls.comparisonRangeStateStore,
        this.parent.timeControls.selectedTimezone,
        this.parent.filters.whereFilter,
        this.parent.filters.dimensionThresholdFilters,
        this.parent.specStore,
        this.parent.timeControls.hasTimeSeries,
        this.specStore,
      ],
      ([
        globalTimeRangeState,
        localTimeRangeState,
        globalComparisonRangeState,
        localComparisonRangeState,
        timeZone,
        whereFilter,
        dtf,
        canvasData,
        hasTimeSeries,
        componentSpec,
      ]) => {
        const metricsViewName = componentSpec["metrics_view"];
        const metricsView = canvasData.data?.metricsViews?.[metricsViewName];
        const dimensions = metricsView?.state?.validSpec?.dimensions ?? [];
        const measures = metricsView?.state?.validSpec?.measures ?? [];

        let timeRange: V1TimeRange = {
          start: globalTimeRangeState?.timeStart,
          end: globalTimeRangeState?.timeEnd,
          timeZone,
        };

        let timeGrain = globalTimeRangeState?.selectedTimeRange?.interval;

        const localShowTimeComparison =
          !!localComparisonRangeState?.showTimeComparison;
        const globalShowTimeComparison =
          !!globalComparisonRangeState?.showTimeComparison;

        let showTimeComparison = globalShowTimeComparison;

        let comparisonTimeRange: V1TimeRange | undefined = {
          start: globalComparisonRangeState?.comparisonTimeStart,
          end: globalComparisonRangeState?.comparisonTimeEnd,
          timeZone,
        };

        let timeRangeState: TimeRangeState | undefined = globalTimeRangeState;
        let comparisonTimeRangeState: ComparisonTimeRangeState | undefined =
          globalComparisonRangeState;

        if (componentSpec?.["time_filters"]) {
          timeRange = {
            start: localTimeRangeState?.timeStart,
            end: localTimeRangeState?.timeEnd,
            timeZone,
          };

          comparisonTimeRange = {
            start: localComparisonRangeState?.comparisonTimeStart,
            end: localComparisonRangeState?.comparisonTimeEnd,
            timeZone,
          };

          showTimeComparison = localShowTimeComparison;

          timeGrain = localTimeRangeState?.selectedTimeRange?.interval;

          timeRangeState = localTimeRangeState;
          comparisonTimeRangeState = localComparisonRangeState;
        }

        // Dimension Filters
        const globalWhere =
          buildValidMetricsViewFilter(whereFilter, dtf, dimensions, measures) ??
          createAndExpression([]);

        let where: V1Expression | undefined = globalWhere;

        if (componentSpec?.["dimension_filters"]) {
          const { expr: componentWhere } = getFiltersFromText(
            componentSpec?.["dimension_filters"] as string,
          );
          where = mergeFilters(globalWhere, componentWhere);
        }

        return {
          timeRange,
          showTimeComparison,
          comparisonTimeRange,
          where,
          timeGrain,
          timeRangeState,
          comparisonTimeRangeState,
          hasTimeSeries,
        };
      },
    );
  }

  private updateYAML(newSpec: T) {
    if (!this.parent.fileArtifact) return;
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent } = this.parent.fileArtifact;

    parsedDocument.setIn(this.pathInYAML, newSpec);

    updateEditorContent(parsedDocument.toString(), false, true);
  }

  setSpec(newSpec: T) {
    if (this.isValid(newSpec)) {
      this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  updateProperty(key: AllKeys<T>, value: T[AllKeys<T>]) {
    const currentSpec = get(this.specStore);

    const newSpec = { ...currentSpec, [key]: value };

    if (value === undefined || value == "") {
      delete newSpec[key];
    }

    // If the metrics_view is changed, clear the time_filters and dimension_filters
    if (key === "metrics_view") {
      if ("time_filters" in newSpec) {
        delete newSpec.time_filters;
      }
      if ("dimension_filters" in newSpec) {
        delete newSpec.dimension_filters;
      }
      if (this.resetParams.length > 0) {
        this.resetParams.forEach((param) => {
          delete newSpec[param];
        });
      }
    }

    if (this.isValid(newSpec)) {
      this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }
}
