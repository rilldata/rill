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
import { derived, writable, type Writable } from "svelte/store";
import { CanvasComponentState } from "./canvas-component";
import { CanvasFilters } from "./canvas-filters";
import { CanvasResolvedSpec } from "./canvas-spec";
import { CanvasTimeControls } from "./canvas-time-control";
import type { GridStack } from "gridstack";

export class CanvasEntity {
  name: string;

  /** Local state store for canvas components */
  components: Map<string, CanvasComponentState>;

  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: CanvasTimeControls;

  /**
   * Dimension and measure filters for the canvas entity
   */
  filters: CanvasFilters;

  /**
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;

  gridstack?: GridStack | null;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string) {
    const validSpecStore: CanvasSpecResponseStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, { queryClient }).subscribe(set),
    );

    this.name = name;

    this.components = new Map();
    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(validSpecStore);
    this.timeControls = new CanvasTimeControls(validSpecStore);
    this.filters = new CanvasFilters(this.spec);
  }

  setSelectedComponentIndex = (index: number | null) => {
    this.selectedComponentIndex.set(index);
  };

  useComponent = (componentName: string) => {
    let componentEntity = this.components.get(componentName);

    if (!componentEntity) {
      componentEntity = new CanvasComponentState(componentName, this.spec);
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
        timeControls.selectedTimeRange,
        timeControls.selectedTimezone,
        timeControls.selectedComparisonTimeRange,
        filters.whereFilter,
        filters.dimensionThresholdFilters,
        dimensionsStore,
        measuresStore,
      ],
      ([
        globalTimeRange,
        timeZone,
        globalComparisonRange,
        whereFilter,
        dtf,
        dimensions,
        measures,
      ]) => {
        // Time Range
        let timeRange: V1TimeRange = {
          start: globalTimeRange?.start?.toISOString(),
          end: globalTimeRange?.end?.toISOString(),
          timeZone,
        };
        if (componentTimeRange) {
          // TODO: update logic
          timeRange = { isoDuration: componentTimeRange, timeZone };
        }
        const timeGrain = globalTimeRange?.interval;

        // Comparison Range
        let comparisonRange: V1TimeRange = {
          start: globalComparisonRange?.start?.toISOString(),
          end: globalComparisonRange?.end?.toISOString(),
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
