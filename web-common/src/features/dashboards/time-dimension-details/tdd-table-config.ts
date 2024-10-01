import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotDataStoreConfig,
  PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, Readable } from "svelte/store";

import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { getPivotConfigKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { TimeRangeString } from "@rilldata/web-common/lib/time/types";

let lastKey: string | undefined = undefined;

export function getTDDConfig(
  ctx: StateManagers,
): Readable<PivotDataStoreConfig> {
  return derived(
    [useMetricsView(ctx), ctx.timeRangeSummaryStore, ctx.dashboardStore],
    ([metricsView, timeRangeSummary, dashboardStore]) => {
      if (
        !metricsView.data?.measures ||
        !metricsView.data?.dimensions ||
        timeRangeSummary.isFetching
      ) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          whereFilter: dashboardStore.whereFilter,
          pivot: dashboardStore.pivot,
          time: {} as PivotTimeConfig,
          comparisonTime: undefined,
          enableComparison: false,
        };
      }

      // This indirection makes sure only one update of dashboard store triggers this
      const timeControl = timeControlStateSelector([
        metricsView,
        timeRangeSummary,
        dashboardStore,
      ]);

      const time: PivotTimeConfig = {
        timeStart: timeControl.selectedTimeRange?.start.toISOString(),
        timeEnd: timeControl.selectedTimeRange?.end.toISOString(),
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView?.data?.timeDimension || "",
      };

      const enableComparison = Boolean(
        timeControl.comparisonTimeStart && !!timeControl.showTimeComparison,
      );

      let comparisonTime: TimeRangeString | undefined = undefined;
      if (enableComparison) {
        comparisonTime = {
          start: timeControl.selectedComparisonTimeRange?.start.toISOString(),
          end: timeControl.selectedComparisonTimeRange?.end.toISOString(),
        };
      }

      const expandedMeasureName = dashboardStore?.tdd.expandedMeasureName;
      const expandedDimensionName = dashboardStore?.selectedComparisonDimension;
      const timeGrain = timeControl.selectedTimeRange?.interval;

      let measureNames: string[] = [];
      let rowDimensionNames: string[] = [];
      let colDimensionNames: string[] = [];

      if (expandedMeasureName) {
        measureNames = [expandedMeasureName];
        if (enableComparison) {
          measureNames.push(
            `${expandedMeasureName}${COMPARISON_DELTA}`,
            `${expandedMeasureName}${COMPARISON_PERCENT}`,
          );
        }
      }
      if (expandedDimensionName) {
        rowDimensionNames = [expandedDimensionName];
      }
      if (timeGrain) {
        colDimensionNames = [`${time.timeDimension}_rill_${timeGrain}`];
      }

      const config: PivotDataStoreConfig = {
        measureNames,
        rowDimensionNames,
        colDimensionNames,
        allMeasures: metricsView.data?.measures || [],
        allDimensions: metricsView.data?.dimensions || [],
        whereFilter: mergeMeasureFilters(dashboardStore),
        pivot: dashboardStore.pivot,
        enableComparison,
        comparisonTime,
        time,
      };

      const currentKey = getPivotConfigKey(config);

      if (lastKey !== currentKey) {
        // Reset rowPage when table config changes
        lastKey = currentKey;
        if (config.pivot.rowPage !== 1) {
          // TODO: Set tdd row page to 1
          config.pivot.rowPage = 1;
        }
      }

      return config;
    },
  );
}
