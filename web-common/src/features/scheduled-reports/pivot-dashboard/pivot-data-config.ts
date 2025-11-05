import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import {
  canEnablePivotComparison,
  getPivotConfigKey,
  splitPivotChips,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotDataStoreConfig,
  type PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { isSimpleMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { PivotStore } from "@rilldata/web-common/features/scheduled-reports/pivot-dashboard/pivot-store.ts";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types.ts";
import { derived, type Readable } from "svelte/store";

export function getPivotConfig(
  metadata: ExploreMetricsViewMetadata,
  filters: Filters,
  timeControls: TimeControls,
  pivotStore: PivotStore,
): Readable<PivotDataStoreConfig> {
  const combinedTimeControlsStore = derived(
    [
      timeControls.getStore(),
      timeControls.timeRangeStateStore,
      timeControls.comparisonRangeStateStore,
      timeControls.minTimeGrain,
    ],
    ([
      timeControlsState,
      timeRangeState,
      comparisonRangeStateState,
      minTimeGrain,
    ]) => {
      return {
        ...(timeControlsState ?? {}),
        ...(timeRangeState ?? {}),
        ...(comparisonRangeStateState ?? {}),
        minTimeGrain,
      };
    },
  );

  let lastKey: string | undefined = undefined;

  return derived(
    [
      metadata.metricsViewSpecQuery,
      metadata.allSimpleMeasures,
      metadata.allDimensions,
      filters.getStore(),
      combinedTimeControlsStore,
      pivotStore.state,
    ],
    ([
      metricsViewSpecResp,
      allSimpleMeasures,
      allDimensions,
      filtersState,
      timeControlsState,
      pivotState,
    ]) => {
      if (metricsViewSpecResp.isPending) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          whereFilter: filtersState.whereFilter,
          pivot: pivotState,
          time: {} as PivotTimeConfig,
          comparisonTime: undefined,
          enableComparison: false,
          searchText: "",
          isFlat: false,
        };
      }

      const metricsViewSpec = metricsViewSpecResp.data ?? {};

      const time: PivotTimeConfig = {
        timeStart: timeControlsState?.timeStart,
        timeEnd: timeControlsState?.timeEnd,
        timeZone: timeControlsState?.selectedTimezone || "UTC",
        timeDimension: metricsViewSpec.timeDimension || "",
        minTimeGrain: timeControlsState.minTimeGrain,
      };

      const enableComparison =
        canEnablePivotComparison(
          pivotState,
          timeControlsState.comparisonTimeStart,
        ) && !!timeControlsState.showTimeComparison;

      let comparisonTime: TimeRangeString | undefined = undefined;
      if (enableComparison) {
        comparisonTime = {
          start: timeControlsState.comparisonTimeStart,
          end: timeControlsState.comparisonTimeEnd,
        };
      }

      const { dimension: colDimensions, measure: colMeasures } =
        splitPivotChips(pivotState.columns);

      const measureNames = colMeasures.flatMap((m) => {
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
      let rowDimensionNames = pivotState.rows.map((d) => {
        if (d.type === PivotChipType.Time) {
          return `${time.timeDimension}_rill_${d.id}`;
        }
        return d.id;
      });

      let colDimensionNames = colDimensions.map((d) => {
        if (d.type === PivotChipType.Time) {
          return `${time.timeDimension}_rill_${d.id}`;
        }
        return d.id;
      });

      const isFlat = pivotState.tableMode === "flat";

      /**
       * For flat table, internally rows have all
       * the dimensions and measures are in columns
       */
      if (isFlat) {
        rowDimensionNames = colDimensionNames;
        colDimensionNames = [];
      }

      const config: PivotDataStoreConfig = {
        measureNames,
        rowDimensionNames,
        colDimensionNames,
        allMeasures: allSimpleMeasures,
        allDimensions,
        whereFilter: mergeDimensionAndMeasureFilters(
          filtersState.whereFilter,
          filtersState.dimensionThresholdFilters,
        ),
        pivot: pivotState,
        enableComparison,
        comparisonTime,
        time,
        searchText: "",
        isFlat,
      };

      const currentKey = getPivotConfigKey(config);

      if (lastKey !== currentKey) {
        // Reset rowPage when pivot config changes
        lastKey = currentKey;
        if (config.pivot.rowPage !== 1) {
          pivotStore.setRowPage(1);
          config.pivot.rowPage = 1;
        }
      }

      return config;
    },
  );
}

export function getAvailableMeasures(
  pivotStore: PivotStore,
  pivotConfigStore: Readable<PivotDataStoreConfig>,
) {
  return derived(
    [pivotStore.columnMeasures, pivotConfigStore],
    ([$columnMeasures, $config]) => {
      return $config.allMeasures
        .filter(
          (m) =>
            isSimpleMeasure(m) && !$columnMeasures.find((c) => c.id === m.name),
        )
        .map((measure) => ({
          id: measure.name || "Unknown",
          title: measure.displayName || measure.name || "Unknown",
          type: PivotChipType.Measure,
          description: measure.description,
        }));
    },
  );
}

export function getAvailableDimensions(
  pivotStore: PivotStore,
  pivotConfigStore: Readable<PivotDataStoreConfig>,
) {
  return derived(
    [pivotStore.state, pivotStore.columnDimensions, pivotConfigStore],
    ([$state, $columnDimensions, $config]) => {
      return $config.allDimensions
        .filter((d) => {
          return !(
            $columnDimensions.find((c) => c.id === d.name) ||
            $state.rows.find((r) => r.id === d.name)
          );
        })
        .map((dimension) => ({
          id: dimension.name || dimension.column || "Unknown",
          title:
            dimension.displayName ||
            dimension.name ||
            dimension.column ||
            "Unknown",
          type: PivotChipType.Dimension,
          description: dimension.description,
        }));
    },
  );
}
