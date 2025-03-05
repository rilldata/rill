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
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { GridStack } from "gridstack";
import { derived, writable, type Readable, type Writable } from "svelte/store";
import { CanvasComponentState } from "./canvas-component";
import { Filters } from "./filters";
import { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";

export class CanvasEntity {
  name: string;

  /** Local state store for canvas components */
  components: Map<string, CanvasComponentState>;

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

  gridstack?: GridStack | null;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponent: Writable<{ row: number; column: number } | null>;

  private specStore: CanvasSpecResponseStore;

  constructor(name: string) {
    this.specStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, {
        retry: 3,
        retryDelay: (attemptIndex) =>
          Math.min(1000 + 1000 * attemptIndex, 5000),
        queryClient,
      }).subscribe(set),
    );

    this.name = name;

    this.components = new Map();
    this.selectedComponent = writable<{ row: number; column: number } | null>(
      null,
    );
    this.spec = new CanvasResolvedSpec(this.specStore);
    this.timeControls = new TimeControls(this.specStore);
    this.filters = new Filters(this.spec);
  }

  setSelectedComponent = (pos: { row: number; column: number } | null) => {
    this.selectedComponent.set(pos);
  };

  useComponent = (componentName: string): CanvasComponentState => {
    let componentEntity = this.components.get(componentName);

    if (!componentEntity) {
      componentEntity = new CanvasComponentState(
        componentName,
        this.specStore,
        this.spec,
      );
      this.components.set(componentName, componentEntity);
    }
    return componentEntity;
  };

  removeComponent = (componentName: string) => {
    this.components.delete(componentName);
  };

  setGridstack(gridstack: GridStack | null) {
    this.gridstack = gridstack;
  }

  /**
   * Helper method to get the time range and where clause for a given metrics view
   * with the ability to override the time range and filter
   */
  componentTimeAndFilterStore = (
    componentName: string,
  ): Readable<TimeAndFilterStore> => {
    const { timeControls, filters, spec, useComponent } = this;

    const componentSpecStore = spec.getComponentResourceFromName(componentName);

    return derived(componentSpecStore, (componentSpec, set) => {
      const metricsViewName = componentSpec?.rendererProperties
        ?.metrics_view as string;

      // if (!metricsViewName) {
      //   throw new Error("Metrics view name is not set for component");
      // }

      const component = useComponent(componentName);
      const dimensionsStore = spec.getDimensionsForMetricView(metricsViewName);
      const measuresStore = spec.getMeasuresForMetricView(metricsViewName);

      return derived(
        [
          timeControls.timeRangeStateStore,
          component.localTimeControls.timeRangeStateStore,
          timeControls.comparisonRangeStateStore,
          component.localTimeControls.comparisonRangeStateStore,
          timeControls.selectedTimezone,
          filters.whereFilter,
          filters.dimensionThresholdFilters,
          dimensionsStore,
          measuresStore,
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

            if (!localShowTimeComparison) {
              showTimeComparison = false;
            }
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
            const componentWhere = component.localFilters.getFiltersFromText(
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
          };
        },
      ).subscribe(set);
    });
  };
}
