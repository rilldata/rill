import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  createAndExpression,
  getValidFilterForMetricView,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { derived, type Writable } from "svelte/store";

// This helper returns a derived store that yields the final timeRange and where clause
export function createTimeAndFilterStore(
  ctx: StateManagers,
  metricsViewName: string,
  {
    timeRangeStore,
    overrideTimeRange,
  }: {
    timeRangeStore: Writable<DashboardTimeControls | undefined>;
    overrideTimeRange?: string;
  },
) {
  const { timeControls, filters, spec } = ctx.canvasEntity;

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
        getValidFilterForMetricView(whereFilter, dtf, dimensions, measures) ??
        createAndExpression([]);

      return { timeRange, where };
    },
  );
}
