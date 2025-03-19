import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
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
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import {
  type Readable,
  type Writable,
  derived,
  get,
  readable,
  writable,
} from "svelte/store";
import type { PivotSpec } from "./";

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

export function getPivotConfig(
  ctx: StateManagers,
  metricsViewName: string,
  tableSpecStore: Readable<PivotSpec>,
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
    },
  );
}

export const usePivotForCanvas = (
  ctx: StateManagers,
  componentName: string,
  metricsViewName: string,
  pivotConfig: Readable<PivotDataStoreConfig>,
) => {
  const cachedEntry = get(tableStoreCache).get(
    `${componentName}-${metricsViewName}`,
  );

  if (cachedEntry) {
    return cachedEntry.store;
  } else {
    clearTableCache(componentName);
  }

  const pivotDashboardContext: PivotDashboardContext = {
    metricsViewName: readable(metricsViewName),
    queryClient: ctx.queryClient,
    enabled: !!ctx.canvasEntity.spec.canvasSpec,
  };

  const pivotDataStore = createPivotDataStore(
    pivotDashboardContext,
    pivotConfig,
  );

  const unsubscribe = pivotDataStore.subscribe(() => {});

  tableStoreCache.update(
    (cache: Map<string, CacheEntry>): Map<string, CacheEntry> => {
      cache.set(`${componentName}-${metricsViewName}`, {
        store: pivotDataStore,
        unsubscribe,
      });
      return cache;
    },
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
    ctx: StateManagers,
    metricsViewName: string,
    tableSpecStore: Readable<TableSpec | PivotSpec>,
    pivotState: Writable<PivotState>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ) => Store,
) {
  const cache = new Map<string, Store>();
  return (
    ctx: StateManagers,
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
export const usePivotConfig = memoizePivotConfig(getPivotConfig);
