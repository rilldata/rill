import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  ComparisonTimeRangeState,
  TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  type V1Expression,
  type V1Resource,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, writable, type Readable } from "svelte/store";
import { Filters } from "./filters";
import { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";
import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
import { createComponent } from "../components/util";
import type { FileArtifact } from "../../entity-management/file-artifact";
import { parseDocument } from "yaml";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import type { CanvasComponentType } from "../components/types";

export class CanvasEntity {
  name: string;

  /** Local state store for canvas components */
  components: Map<string, BaseCanvasComponent>;

  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: TimeControls;

  /**
   * Dimension and measure filters for the canvas entity
   */
  filters: Filters;

  /**
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponent = writable<string | null>(null);

  fileArtifact: FileArtifact | undefined;
  parsedContent: Readable<ReturnType<typeof parseDocument>>;

  specStore: CanvasSpecResponseStore;

  constructor(name: string) {
    const instanceId = get(runtime).instanceId;
    this.specStore = useCanvas(instanceId, name, {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 + 1000 * attemptIndex, 5000),
      queryClient,
    });

    this.name = name;

    this.components = new Map();

    this.spec = new CanvasResolvedSpec(this.specStore);
    this.timeControls = new TimeControls(this.specStore);
    this.filters = new Filters(this.spec);

    this.specStore.subscribe((spec) => {
      const filePath = spec.data?.filePath;
      if (!filePath || this.fileArtifact) {
        return;
      }
      const fileArtifact = fileArtifacts.getFileArtifact(filePath);

      if (!fileArtifact) {
        return;
      }
      this.fileArtifact = fileArtifact;

      if (this.parsedContent) {
        return;
      }

      this.parsedContent = derived(
        fileArtifact.editorContent,
        (editorContent) => {
          const parsed = parseDocument(editorContent ?? "");
          return parsed;
        },
      );

      const components = spec.data?.components;
      const rows = spec.data?.canvas?.rows ?? [];

      rows.forEach((row, rowIndex) => {
        const items = row.items ?? [];
        items.forEach((item, columnIndex) => {
          const componentName = item.component;
          if (!componentName) return;

          const resource = components?.[componentName];
          if (!resource) return;

          this.initializeOrUpdateComponent(resource, rowIndex, columnIndex);
        });
      });
    });
  }

  setSelectedComponent = (id: string | null) => {
    this.selectedComponent.set(id);
  };

  initializeOrUpdateComponent = (
    resource: V1Resource,
    row: number,
    column: number,
  ) => {
    const id = resource.meta?.name?.name as string;
    const component = this.components.get(id);
    const path = constructPath(
      row,
      column,
      resource.component?.state?.validSpec?.renderer as CanvasComponentType,
    );
    if (component) {
      component.update(resource, path);
    } else {
      this.components.set(id, createComponent(resource, this, path));
    }
  };

  // useComponent = (componentName: string): BaseCanvasComponent => {
  //   let componentEntity = this.components.get(componentName);

  //   if (!componentEntity) {
  //     const canvasResponse = get(this.specStore).data;

  //     const components = canvasResponse?.components;
  //     const rows = canvasResponse?.canvas?.rows ?? [];

  //     const [row, column] = rows.reduce(
  //       (acc, row, index) => {
  //         const items = row.items ?? [];
  //         const item = items.find((item) => item.component === componentName);
  //         if (item) {
  //           return [index, items.indexOf(item)];
  //         }
  //         return acc;
  //       },
  //       [-1, -1],
  //     );
  //     const resource = components?.[componentName];
  //     if (!resource) {
  //       throw new Error(`Component ${componentName} not found in canvas spec`);
  //     }

  //     const path = constructPath(
  //       row,
  //       column,
  //       resource.component?.state?.validSpec?.renderer as CanvasComponentType,
  //     );

  //     componentEntity = createComponent(resource, this, path);
  //     this.components.set(componentName, componentEntity);
  //   }
  //   return componentEntity;
  // };

  removeComponent = (componentName: string) => {
    this.components.delete(componentName);
  };

  /**
   * Helper method to get the time range and where clause for a given metrics view
   * with the ability to override the time range and filter
   */
  componentTimeAndFilterStore = (
    componentName: string,
  ): Readable<TimeAndFilterStore> => {
    const { timeControls, filters, spec } = this;

    const componentSpecStore = spec.getComponentResourceFromName(componentName);

    return derived(componentSpecStore, (componentSpec, set) => {
      const metricsViewName = componentSpec?.rendererProperties
        ?.metrics_view as string;

      // if (!metricsViewName) {
      //   throw new Error("Metrics view name is not set for component");
      // }

      const component = this.components.get(componentName);
      if (!component) {
        throw new Error(
          `Time and filter store: Component ${componentName} not found`,
        );
      }
      const dimensionsStore = spec.getDimensionsForMetricView(metricsViewName);
      const measuresStore = spec.getMeasuresForMetricView(metricsViewName);

      return derived(
        [
          timeControls.timeRangeStateStore,
          component.state.localTimeControls.timeRangeStateStore,
          timeControls.comparisonRangeStateStore,
          component.state.localTimeControls.comparisonRangeStateStore,
          timeControls.selectedTimezone,
          filters.whereFilter,
          filters.dimensionThresholdFilters,
          dimensionsStore,
          measuresStore,
          timeControls.hasTimeSeries,
        ],
        ([
          globalTimeRangeState,
          localTimeRangeState,
          globalComparisonRangeState,
          localComparisonRangeState,
          timeZone,
          whereFilter,
          dtf,
          dimensions,
          measures,
          hasTimeSeries,
        ]) => {
          // Time Filters
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

          if (componentSpec?.rendererProperties?.time_filters) {
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
            buildValidMetricsViewFilter(
              whereFilter,
              dtf,
              dimensions,
              measures,
            ) ?? createAndExpression([]);

          let where: V1Expression | undefined = globalWhere;

          if (componentSpec?.rendererProperties?.dimension_filters) {
            const { expr: componentWhere } =
              component.state.localFilters.getFiltersFromText(
                componentSpec.rendererProperties.dimension_filters as string,
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
      ).subscribe(set);
    });
  };
}

export type ComponentPath = [
  "rows",
  number,
  "items",
  number,
  CanvasComponentType,
];

function constructPath(
  row: number,
  column: number,
  type: CanvasComponentType,
): ComponentPath {
  return ["rows", row, "items", column, type];
}
