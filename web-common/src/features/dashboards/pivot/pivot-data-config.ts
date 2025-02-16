import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { allDimensions } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimensions";
import { allMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  dimensionSearchText,
  metricsExplorerStore,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { timeControlStateSelector } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import { type Readable, derived } from "svelte/store";
import { canEnablePivotComparison, getPivotConfigKey } from "./pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotDataStoreConfig,
  type PivotTimeConfig,
} from "./types";

let lastKey: string | undefined = undefined;

/**
 * Extract out config relevant to pivot from dashboard and meta store
 */
export function getPivotConfig(
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
          isFlat: false,
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
        timeStart: timeControl.timeStart,
        timeEnd: timeControl.timeEnd,
        timeZone: dashboardStore?.selectedTimezone || "UTC",
        timeDimension: metricsView.timeDimension || "",
      };

      const enableComparison =
        canEnablePivotComparison(
          dashboardStore.pivot,
          timeControl.comparisonTimeStart,
        ) && !!timeControl.showTimeComparison;

      let comparisonTime: TimeRangeString | undefined = undefined;
      if (enableComparison) {
        comparisonTime = {
          start: timeControl.comparisonTimeStart,
          end: timeControl.comparisonTimeEnd,
        };
      }

      const measureNames = dashboardStore.pivot.columns.measure.flatMap((m) => {
        const measureName = m.id;
        const group = [measureName];

        if (enableComparison) {
          group.push(
            `${measureName}${COMPARISON_DELTA}`,
            `${measureName}${COMPARISON_PERCENT}`,
          );
        }
        return group;
      });

      // This is temporary until we have a better way to handle time grains
      const rowDimensionNames = dashboardStore.pivot.rows.dimension.map((d) => {
        if (d.type === PivotChipType.Time) {
          return `${time.timeDimension}_rill_${d.id}`;
        }
        return d.id;
      });

      let colDimensionNames = dashboardStore.pivot.columns.dimension.map(
        (d) => {
          if (d.type === PivotChipType.Time) {
            return `${time.timeDimension}_rill_${d.id}`;
          }
          return d.id;
        },
      );

      // const isFlat = dashboardStore.pivot.rowJoinType === "flat";
      const isFlat = false;
      /**
       * For flat table, internally rows have all
       * the dimensions and measures are in columns
       */
      if (isFlat) colDimensionNames = [];

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
        isFlat,
      };

      const currentKey = getPivotConfigKey(config);

      if (lastKey !== currentKey) {
        // Reset rowPage when pivot config changes
        lastKey = currentKey;
        if (config.pivot.rowPage !== 1) {
          metricsExplorerStore.setPivotRowPage(dashboardStore.name, 1);
          config.pivot.rowPage = 1;
        }
      }

      return config;
    },
  );
}
