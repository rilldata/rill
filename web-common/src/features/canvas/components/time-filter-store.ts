import {
  useAllDimensionFromMetric,
  useAllMeasuresFromMetric,
} from "@rilldata/web-common/features/canvas/components/selectors";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { getValidFilterForMetricView } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { derived, type Writable } from "svelte/store";

// This helper returns a derived store that yields the final timeRange and where clause
export function createTimeAndFilterStore(
  ctx: StateManagers,
  instanceId: string,
  metricsViewName: string,
  {
    timeRangeStore,
    overrideTimeRange,
  }: {
    timeRangeStore: Writable<DashboardTimeControls | undefined>;
    overrideTimeRange?: string;
  },
) {
  const { timeControls, filters } = ctx.canvasEntity;
  const dimensionsQuery = useAllDimensionFromMetric(
    instanceId,
    metricsViewName,
  );

  const measuresQuery = useAllMeasuresFromMetric(instanceId, metricsViewName);
  return derived(
    [
      timeRangeStore,
      timeControls.selectedTimezone,
      filters.whereFilter,
      filters.dimensionThresholdFilters,
      dimensionsQuery,
      measuresQuery,
    ],
    ([timeRangeVal, timeZone, whereFilter, dtf, dimensions, measures]) => {
      // 1. Build up the final V1TimeRange
      let timeRange: V1TimeRange = {
        start: timeRangeVal?.start?.toISOString(),
        end: timeRangeVal?.end?.toISOString(),
        timeZone,
      };
      if (overrideTimeRange) {
        timeRange = { isoDuration: overrideTimeRange, timeZone };
      }

      // 2. Get the valid "where" expression
      const where = getValidFilterForMetricView(
        whereFilter,
        dtf,
        dimensions.data || [],
        measures.data || [],
      );

      return { timeRange, where };
    },
  );
}
