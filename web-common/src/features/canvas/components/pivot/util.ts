import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
import {
  canEnablePivotComparison,
  getPivotConfigKey,
  getTimeGrainFromDimension,
  isTimeDimension,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotDashboardContext,
  type PivotDataStoreConfig,
  type PivotState,
  type PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type {
  V1Expression,
  V1MetricsViewSpec,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { type Readable, type Writable, derived, writable } from "svelte/store";
import type { CanvasEntity } from "../../stores/canvas-entity";
import type { PivotSpec, TableSpec } from "./";

type CacheEntry = {
  store: ReturnType<typeof createPivotDataStore>;
  unsubscribe: () => void;
};

const tableStoreCache = writable<Map<string, CacheEntry>>(new Map());
export function clearTableCache(componentName?: string) {
  tableStoreCache.update(
    (cache: Map<string, CacheEntry>): Map<string, CacheEntry> => {
      if (!componentName) {
        // Clear all cache entries if componentName is undefined
        for (const entry of cache.values()) {
          entry.unsubscribe();
        }
        cache.clear();
      } else {
        // Clear only entries matching componentName
        for (const [key, entry] of cache.entries()) {
          if (key.startsWith(componentName)) {
            entry.unsubscribe();
            cache.delete(key);
          }
        }
      }
      return cache;
    },
  );
}
let lastKey: string | undefined = undefined;

export function createPivotConfig(
  canvas: CanvasEntity,
  tableSpecStore: Readable<PivotSpec | TableSpec>,
  pivotState: Writable<PivotState>,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<PivotDataStoreConfig> {
  return derived(
    [canvas.specStore, tableSpecStore, pivotState, timeAndFilterStore],
    ([$canvasData, $tableSpec, $pivotState, $timeAndFilterStore]) => {
      const { timeRange, comparisonTimeRange, where } = $timeAndFilterStore;
      const metricsViewName = $tableSpec.metrics_view;
      const metricsView =
        $canvasData?.data?.metricsViews[metricsViewName]?.state?.validSpec ??
        {};

      return "columns" in $tableSpec
        ? processFlat(
            $tableSpec,
            $pivotState,
            where,
            metricsView,
            $timeAndFilterStore,
            comparisonTimeRange,
            pivotState,
            timeRange,
          )
        : processPivot(
            $tableSpec,
            $pivotState,
            where,
            metricsView,
            $timeAndFilterStore,
            comparisonTimeRange,
            pivotState,
            timeRange,
          );
    },
  );
}

export function processPivot(
  $tableSpec: PivotSpec,
  $pivotState: PivotState,
  where: V1Expression | undefined,
  metricsView: V1MetricsViewSpec | undefined,
  $timeAndFilterStore: TimeAndFilterStore,
  comparisonTimeRange: V1TimeRange | undefined,
  pivotState: Writable<PivotState>,
  timeRange: V1TimeRange,
) {
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
      isFlat: false,
    };
  }

  const enableComparison =
    canEnablePivotComparison($pivotState, comparisonTimeRange?.start) &&
    $timeAndFilterStore.showTimeComparison;

  const config: PivotDataStoreConfig = {
    measureNames: ($tableSpec?.measures || []).flatMap((name) => {
      const group = [name];
      if (enableComparison) {
        group.push(
          `${name}${COMPARISON_DELTA}`,
          `${name}${COMPARISON_PERCENT}`,
        );
      }
      return group;
    }),
    rowDimensionNames: $tableSpec?.row_dimensions || [],
    colDimensionNames: $tableSpec?.col_dimensions || [],
    allMeasures: metricsView?.measures || [],
    allDimensions: metricsView?.dimensions || [],
    whereFilter: where ?? createAndExpression([]),
    searchText: "",
    isFlat: false,
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
}

export function processFlat(
  $tableSpec: TableSpec,
  $pivotState: PivotState,
  where: V1Expression | undefined,
  metricsView: V1MetricsViewSpec | undefined,
  $timeAndFilterStore: TimeAndFilterStore,
  comparisonTimeRange: V1TimeRange | undefined,
  pivotState: Writable<PivotState>,
  timeRange: V1TimeRange,
) {
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
}

export const usePivotForCanvas = (
  canvas: CanvasEntity,
  metricsViewStore: Readable<string>,
  pivotConfig: Readable<PivotDataStoreConfig>,
) => {
  const pivotDashboardContext: PivotDashboardContext = {
    metricsViewName: metricsViewStore,
    queryClient: queryClient,
    enabled: !!canvas.spec,
  };

  const pivotDataStore = createPivotDataStore(
    pivotDashboardContext,
    pivotConfig,
  );

  return pivotDataStore;
};

export function tableFieldMapper(
  fields: string[],
  metricViewSpec: V1MetricsViewSpec | undefined,
) {
  const timeDimension = metricViewSpec?.timeDimension;
  const measures = metricViewSpec?.measures?.map((m) => m.name as string) || [];
  return fields.map((field) => {
    if (timeDimension && isTimeDimension(field, timeDimension)) {
      const grain = getTimeGrainFromDimension(field);
      return {
        id: grain,
        title: `Time ${grain}`,
        type: PivotChipType.Time,
      };
    }
    if (measures.includes(field)) {
      return {
        id: field,
        title: field,
        type: PivotChipType.Measure,
      };
    }
    return {
      id: field,
      title: field,
      type: PivotChipType.Dimension,
    };
  });
}

export function memoizePivotConfig<
  Store extends Readable<PivotDataStoreConfig>,
>(
  storeGetter: (
    ctx: CanvasStore,
    metricsViewName: string,
    tableSpecStore: Readable<TableSpec | PivotSpec>,
    pivotState: Writable<PivotState>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ) => Store,
) {
  const cache = new Map<string, Store>();
  return (
    ctx: CanvasStore,
    metricsViewName: string,
    tableSpecStore: Readable<TableSpec | PivotSpec>,
    pivotState: Writable<PivotState>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): Store => {
    return derived(tableSpecStore, ($tableSpec, set) => {
      const key = JSON.stringify($tableSpec);
      let store = cache.get(key);
      if (!store) {
        store = storeGetter(
          ctx,
          metricsViewName,
          tableSpecStore,
          pivotState,
          timeAndFilterStore,
        );
        cache.set(key, store);
      }
      return store.subscribe(set);
    }) as Store;
  };
}
