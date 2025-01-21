import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, writable, type Writable } from "svelte/store";
import { CanvasFilters } from "./canvas-filters";
import { CanvasResolvedSpec } from "./canvas-spec";
import { CanvasTimeControls } from "./canvas-time-control";

export class CanvasEntity {
  name: string;

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

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string) {
    const validSpecStore: CanvasSpecResponseStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, { queryClient }).subscribe(set),
    );

    this.name = name;

    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(validSpecStore);
    this.timeControls = new CanvasTimeControls(validSpecStore);
    this.filters = new CanvasFilters(this.spec);
  }

  setSelectedComponentIndex(index: number | null) {
    this.selectedComponentIndex.set(index);
  }

  /**
   * Helper method to get the time range and where clause for a given metrics view
   * with the ability to override the time range and filter
   */
  createTimeAndFilterStore(
    metricsViewName: string,
    {
      timeRangeStore,
      overrideTimeRange,
    }: {
      timeRangeStore: Writable<DashboardTimeControls | undefined>;
      overrideTimeRange?: string;
    },
  ) {
    const { timeControls, filters, spec } = this;

    const dimensionsStore = spec.getDimensionsForMetricView(metricsViewName);
    const measuresStore = spec.getMeasuresForMetricView(metricsViewName);

    return derived(
      [
        timeRangeStore,
        timeControls.selectedTimezone,
        filters.whereFilter,
        filters.dimensionThresholdFilters,
        dimensionsStore,
        measuresStore,
      ],
      ([timeRangeVal, timeZone, whereFilter, dtf, dimensions, measures]) => {
        let timeRange: V1TimeRange = {
          start: timeRangeVal?.start?.toISOString(),
          end: timeRangeVal?.end?.toISOString(),
          timeZone,
        };
        if (overrideTimeRange) {
          timeRange = { isoDuration: overrideTimeRange, timeZone };
        }

        const where =
          buildValidMetricsViewFilter(whereFilter, dtf, dimensions, measures) ??
          createAndExpression([]);

        return { timeRange, where };
      },
    );
  }
}
