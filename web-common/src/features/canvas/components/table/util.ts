import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import {
  canEnablePivotComparison,
  getPivotConfigKey,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDataStoreConfig,
  type PivotState,
  type PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type Readable, type Writable, derived } from "svelte/store";
import type { TableSpec } from "./";

let lastKey: string | undefined = undefined;

export function getTableConfig(
  ctx: StateManagers,
  metricsViewName: string,
  tableSpecStore: Readable<TableSpec>,
  pivotState: Writable<PivotState>,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<PivotDataStoreConfig> {
  const {
    canvasEntity: {
      spec: { getMetricsViewFromName },
    },
  } = ctx;

  return derived(
    [
      getMetricsViewFromName(metricsViewName),
      tableSpecStore,
      pivotState,
      timeAndFilterStore,
    ],
    ([metricsView, $tableSpec, $pivotState, $timeAndFilterStore]) => {
      const { timeRange, comparisonTimeRange, where } = $timeAndFilterStore;

      if (!$tableSpec) {
        return {
          measureNames: [],
          rowDimensionNames: [],
          colDimensionNames: [],
          allMeasures: [],
          allDimensions: [],
          whereFilter: where ?? createAndExpression([]),
          pivot: $pivotState,
          time: {} as PivotTimeConfig,
          comparisonTime: undefined,
          enableComparison: false,
          searchText: "",
          isFlat: true,
        };
      }

      const columns = $tableSpec?.columns || [];
      const allMeasureNames =
        metricsView?.measures?.map((m) => m.name as string) || [];

      const measures = columns.filter((c) => allMeasureNames.includes(c)) || [];
      const dimensions = columns.filter((c) => !measures.includes(c)) || [];

      const enableComparison =
        canEnablePivotComparison($pivotState, comparisonTimeRange?.start) &&
        $timeAndFilterStore.showTimeComparison;

      const config: PivotDataStoreConfig = {
        measureNames: (measures || []).flatMap((name) => {
          const group = [name];
          if (enableComparison) {
            group.push(
              `${name}${COMPARISON_DELTA}`,
              `${name}${COMPARISON_PERCENT}`,
            );
          }
          return group;
        }),
        rowDimensionNames: dimensions || [],
        colDimensionNames: [],
        allMeasures: metricsView?.measures || [],
        allDimensions: metricsView?.dimensions || [],
        whereFilter: where ?? createAndExpression([]),
        searchText: "",
        isFlat: true,
        pivot: $pivotState,
        enableComparison,
        comparisonTime: {
          start: comparisonTimeRange?.start,
          end: comparisonTimeRange?.end,
        },
        time: {
          timeStart: timeRange?.start,
          timeEnd: timeRange?.end,
          timeZone: timeRange.timeZone || "UTC",
          timeDimension: metricsView?.timeDimension || "",
        },
      };

      const currentKey = getPivotConfigKey(config);

      if (lastKey !== currentKey) {
        // Reset rowPage when pivot config changes
        lastKey = currentKey;
        if (config.pivot.rowPage !== 1) {
          pivotState.update((state) => ({
            ...state,
            rowPage: 1,
          }));
        }
      }

      return config;
    },
  );
}
