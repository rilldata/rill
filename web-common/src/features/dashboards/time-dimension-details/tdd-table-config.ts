import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDataStoreConfig,
  type PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { derived, type Readable } from "svelte/store";

import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { getPivotConfigKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { allDimensions } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimensions";
import { allMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import { dimensionSearchText } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { type TimeRangeString } from "@rilldata/web-common/lib/time/types";

let lastKey: string | undefined = undefined;

export function getTDDConfig(
  ctx: StateManagers,
): Readable<PivotDataStoreConfig> {
  return derived(
    [
      ctx.validSpecStore,
      ctx.timeRangeSummaryStore,
      ctx.dashboardStore,
      dimensionSearchText,
    ],
    ([validSpec, timeRangeSummary, dashboardStore, searchText]) => {
      if (
        !validSpec?.data?.metricsView ||
        !validSpec?.data?.explore ||
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
          searchText,
        };
      }
      const { metricsView, explore } = validSpec.data;

      // This indirection makes sure only one update of dashboard store triggers this
      const timeControl = timeControlStateSelector([
        metricsView,
        explore,
        timeRangeSummary,
        dashboardStore,
      ]);

      const time: PivotTimeConfig = {
        timeStart: timeControl.selectedTimeRange?.start.toISOString(),
        timeEnd: timeControl.selectedTimeRange?.end.toISOString(),
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView.timeDimension || "",
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
        allMeasures: allMeasures({
          validMetricsView: metricsView,
          validExplore: explore,
        }),
        allDimensions: allDimensions({
          validMetricsView: metricsView,
          validExplore: explore,
        }),
        whereFilter: mergeMeasureFilters(dashboardStore),
        pivot: dashboardStore.pivot,
        enableComparison,
        comparisonTime,
        time,
        searchText,
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
