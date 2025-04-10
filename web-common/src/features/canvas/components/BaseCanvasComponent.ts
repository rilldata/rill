import type {
  CanvasComponentType,
  ComponentSize,
  ComponentSpec,
} from "@rilldata/web-common/features/canvas/components/types";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  V1Expression,
  V1Resource,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { derived, get, writable, type Writable } from "svelte/store";
import { CanvasComponentState } from "../stores/canvas-component";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import type {
  ComparisonTimeRangeState,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "../../dashboards/stores/filter-utils";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";
import type { ComponentType, SvelteComponent } from "svelte";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copoy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Local filters and local time controls (will be moved out of class)
  state: CanvasComponentState<T>;
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
    this.state = new CanvasComponentState(
      this.id,
      this.parent.specStore,
      this.parent.spec,
      this.specStore,
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
        this.state.localTimeControls.timeRangeStateStore,
        this.parent.timeControls.comparisonRangeStateStore,
        this.state.localTimeControls.comparisonRangeStateStore,
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
          const { expr: componentWhere } =
            this.state.localFilters.getFiltersFromText(
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
