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
  TimeAndFilterStore,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import type {
  CanvasEntity,
  ComponentPath,
  SearchParamsStore,
} from "../stores/canvas-entity";
import { TimeState } from "../stores/time-state";
import type { FilterState } from "../stores/metrics-view-filter";
import type { Readable } from "svelte/motion";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copoy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Widget specific dimension and measure filters
  localFilters: FilterState;
  // Widget specific time filters
  localTimeControls: TimeState;

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

  metricsViewName: string;

  constructor(
    resource: V1Resource,
    public parent: CanvasEntity,
    path: ComponentPath,
    public defaultSpec: T,
  ) {
    const yamlSpec = resource.component?.state?.validSpec?.rendererProperties;

    const mergedSpec = { ...defaultSpec, ...yamlSpec };
    this.metricsViewName = mergedSpec["metrics_view"] as string;
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
        set: (map: Map<string, string | undefined>) => {
          const searchParams = get(store);

          map.forEach((value, key) => {
            if (value === undefined || value === null || value === "") {
              searchParams.delete(key);
            } else {
              searchParams.set(key, value);
            }
          });

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

    this.localTimeControls = this.parent.timeManager.createLocalTimeState(
      this.id,
      yamlTimeFilterStore,
    );

    this.localFilters = this.parent.filterManager.createLocalFilterStore(
      this.metricsViewName,
    );

    this.specStore.subscribe((spec) => {
      this.localFilters.onFilterStringChange(
        spec["dimension_filters"] as string,
      );
      this.localTimeControls.onUrlChange(
        new URLSearchParams(spec?.["time_filters"] ?? ""),
      );
    });
  }

  update(resource: V1Resource, path: ComponentPath) {
    const yamlSpec = resource.component?.state?.validSpec
      ?.rendererProperties as T;
    this.resource.set(resource);
    this.pathInYAML = path;
    this.specStore.set(yamlSpec);
  }

  // This will be deprecated eventually - bgh
  get timeAndFilterStore(): Readable<TimeAndFilterStore> {
    return derived(
      [
        this.parent.timeManager.state.interval,
        this.parent.timeManager.state.grainStore,
        this.parent.timeManager.state.showTimeComparisonStore,
        this.parent.timeManager.state.comparisonRangeStore,
        this.parent.timeManager.state.comparisonIntervalStore,
        this.parent.timeManager.state.timeZoneStore,
        this.localTimeControls.interval,
        this.localTimeControls.comparisonIntervalStore,
        this.localTimeControls.showTimeComparisonStore,
        this.localTimeControls.grainStore,
        this.localTimeControls.comparisonRangeStore,
        this.parent.filterManager.metricsViewFilters,
        this.parent.specStore,
        this.parent.timeManager.hasTimeSeriesMap,
        this.specStore,
      ],
      (
        [
          globalInterval,
          globalGrainStore,
          globalShowTimeComparison,
          globalComparisonRange,
          globalComparisonInterval,
          timeZone,
          localInterval,
          localComparisonInterval,
          localShowTimeComparison,
          localGrainStore,
          localComparisonRange,
          metricsViewFilters,
          canvasData,
          hasTimeSeriesMap,
          componentSpec,
        ],
        set,
      ) => {
        const hasTimeSeries =
          hasTimeSeriesMap.get(this.metricsViewName) ?? false;

        const mvFilters = metricsViewFilters.get(this.metricsViewName);

        if (!mvFilters) {
          set({
            timeRange: { start: "", end: "", timeZone: "UTC" },
            where: undefined,
            timeGrain: undefined,
            hasTimeSeries: false,
            comparisonTimeRange: undefined,
            timeRangeState: undefined,
            comparisonTimeRangeState: undefined,
            showTimeComparison: false,
          });
          return;
        }

        derived([mvFilters.parsed], ([parsedFilter]) => {
          const metricsViewName = componentSpec["metrics_view"];
          const metricsView = canvasData.data?.metricsViews?.[metricsViewName];
          const dimensions = metricsView?.state?.validSpec?.dimensions ?? [];
          const measures = metricsView?.state?.validSpec?.measures ?? [];

          let timeRange: V1TimeRange = {
            start: globalInterval?.start.toISO(),
            end: globalInterval?.end.toISO(),
            timeZone,
          };

          let timeGrain = globalGrainStore;

          let showTimeComparison = globalShowTimeComparison;

          let comparisonTimeRange: V1TimeRange | undefined = {
            start: globalComparisonInterval?.start.toISO(),
            end: globalComparisonInterval?.end.toISO(),
            timeZone,
          };

          let timeRangeState: TimeRangeState | undefined = {
            timeStart: globalInterval?.start.toISO(),
            timeEnd: globalInterval?.end.toISO(),
          };
          let comparisonTimeRangeState: ComparisonTimeRangeState | undefined =
            globalComparisonInterval && {
              comparisonTimeStart: globalComparisonInterval.start.toISO(),
              comparisonTimeEnd: globalComparisonInterval.end.toISO(),
              selectedComparisonTimeRange: {
                start: globalComparisonInterval.start.toJSDate(),
                end: globalComparisonInterval.end.toJSDate(),
                name: globalComparisonRange,
              },
            };

          if (componentSpec?.["time_filters"]) {
            timeRange = {
              start: localInterval?.start.toISO(),
              end: localInterval?.end.toISO(),
              timeZone,
            };

            comparisonTimeRange = {
              start: localComparisonInterval?.start.toISO(),
              end: localComparisonInterval?.end.toISO(),
              timeZone,
            };

            showTimeComparison = localShowTimeComparison;

            timeGrain = localGrainStore ?? globalGrainStore;

            const localTimeRangeState: TimeRangeState = {
              timeStart: localInterval?.start.toISO(),
              timeEnd: localInterval?.end.toISO(),
            };
            const localComparisonRangeState:
              | ComparisonTimeRangeState
              | undefined = localComparisonInterval && {
              comparisonTimeStart: localComparisonInterval.start.toISO(),
              comparisonTimeEnd: localComparisonInterval.end.toISO(),
              selectedComparisonTimeRange: {
                start: localComparisonInterval.start.toJSDate(),
                end: localComparisonInterval.end.toJSDate(),
                name: localComparisonRange,
              },
            };

            timeRangeState = localTimeRangeState;
            comparisonTimeRangeState = localComparisonRangeState;
          }

          // Dimension Filters
          const globalWhere =
            buildValidMetricsViewFilter(
              parsedFilter.where,
              parsedFilter.dimensionThresholdFilters,
              dimensions,
              measures,
            ) ?? createAndExpression([]);

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
        }).subscribe(set);
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
