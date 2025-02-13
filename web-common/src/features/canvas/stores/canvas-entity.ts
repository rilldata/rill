import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  V1Operation,
  type V1Expression,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { GridStack } from "gridstack";
import { derived, writable, type Writable } from "svelte/store";
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
  selectedComponentIndex: Writable<number | null>;

  private specStore: CanvasSpecResponseStore;

  constructor(name: string) {
    this.specStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, { queryClient }).subscribe(set),
    );

    this.name = name;

    this.components = new Map();
    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(this.specStore);
    this.timeControls = new TimeControls(this.specStore);
    this.filters = new Filters(this.spec);
  }

  setSelectedComponentIndex = (index: number | null) => {
    this.selectedComponentIndex.set(index);
  };

  useComponent = (componentName: string) => {
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

  setGridstack(gridstack: GridStack | null) {
    this.gridstack = gridstack;
  }

  /**
   * Helper method to get the time range and where clause for a given metrics view
   * with the ability to override the time range and filter
   */
  createTimeAndFilterStore = (
    metricsViewName: string,
    {
      componentTimeRange,
      componentComparisonRange,
      componentFilter,
    }: {
      timeRangeStore?: Writable<DashboardTimeControls | undefined>;
      componentTimeRange?: string;
      componentComparisonRange?: string;
      componentFilter?: string;
    },
  ) => {
    const { timeControls, filters, spec } = this;

    const dimensionsStore = spec.getDimensionsForMetricView(metricsViewName);
    const measuresStore = spec.getMeasuresForMetricView(metricsViewName);

    return derived(
      [
        timeControls.timeRangeStateStore,
        timeControls.comparisonRangeStateStore,
        timeControls.selectedTimezone,
        filters.whereFilter,
        filters.dimensionThresholdFilters,
        dimensionsStore,
        measuresStore,
      ],
      ([
        globalTimeRangeState,
        globalComparisonRangeState,
        timeZone,
        whereFilter,
        dtf,
        dimensions,
        measures,
      ]) => {
        // Time Range
        let timeRange: V1TimeRange = {
          start: globalTimeRangeState?.timeStart,
          end: globalTimeRangeState?.timeEnd,
          timeZone,
        };
        if (componentTimeRange) {
          // TODO: update logic
          timeRange = { isoDuration: componentTimeRange, timeZone };
        }
        const timeGrain = globalTimeRangeState?.selectedTimeRange?.interval;

        // Comparison Range
        let comparisonRange: V1TimeRange = {
          start: globalComparisonRangeState?.comparisonTimeStart,
          end: globalComparisonRangeState?.comparisonTimeEnd,
          timeZone,
        };
        if (componentComparisonRange) {
          // TODO: update logic
          comparisonRange = { isoDuration: componentComparisonRange, timeZone };
        }

        // Dimension Filters
        const globalWhere =
          buildValidMetricsViewFilter(whereFilter, dtf, dimensions, measures) ??
          createAndExpression([]);

        let where: V1Expression | undefined = globalWhere;

        if (componentFilter) {
          let expr = convertFilterParamToExpression(componentFilter);
          if (
            expr?.cond?.op !== V1Operation.OPERATION_AND &&
            expr?.cond?.op !== V1Operation.OPERATION_OR
          ) {
            expr = createAndExpression([expr]);
          }
          where = mergeFilters(globalWhere, expr);
        }

        return { timeRange, comparisonRange, where, timeGrain };
      },
    );
  };
}
