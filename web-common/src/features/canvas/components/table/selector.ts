import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
import { canEnablePivotComparison } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDashboardContext,
  type PivotDataStoreConfig,
  type PivotState,
  type PivotTimeConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type Readable, derived, readable, writable } from "svelte/store";

export const pivotState = writable<PivotState>({
  active: true,
  columns: { measure: [], dimension: [] },
  rows: { dimension: [] },
  expanded: {},
  sorting: [],
  columnPage: 1,
  rowPage: 1,
  enableComparison: false,
  rowJoinType: "nest",
  activeCell: null,
});

export function getTableConfig(
  ctx: StateManagers,
  metricsViewName: string,
  tableSpecStore: Readable<TableSpec>,
  pivotState: Readable<PivotState>,
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

      const enableComparison = canEnablePivotComparison(
        $pivotState,
        comparisonTimeRange?.start,
      );

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

      return config;
    },
  );
}

export function validateTableSchema(
  ctx: StateManagers,
  tableSpec: TableSpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view } = tableSpec;
  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      const measures = tableSpec.measures || [];
      const rowDimensions = tableSpec.row_dimensions || [];
      const colDimensions = tableSpec.col_dimensions || [];

      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

      if (!measures.length && !rowDimensions.length && !colDimensions.length) {
        return {
          isValid: false,
          error: "Select at least one measure or dimension for the table",
        };
      }
      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure(s) "${invalidMeasures}" selected for the table`,
        };
      }

      const validateDimensionsRes = validateDimensions(
        metricsView,
        rowDimensions.concat(colDimensions),
      );

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) "${invalidDimensions}" selected for the table`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}

export const usePivotForCanvas = (() => {
  const cache = new Map<string, ReturnType<typeof createPivotDataStore>>();

  return (
    ctx: StateManagers,
    metricsViewName: string,
    tableSpecStore: Readable<TableSpec>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ) => {
    if (cache.has(metricsViewName)) {
      return cache.get(metricsViewName)!;
    }

    const pivotConfig = getTableConfig(
      ctx,
      metricsViewName,
      tableSpecStore,
      pivotState,
      timeAndFilterStore,
    );

    const pivotDashboardContext: PivotDashboardContext = {
      metricsViewName: readable(metricsViewName),
      queryClient: ctx.queryClient,
      enabled: !!ctx.canvasEntity.spec.canvasSpec,
    };

    const pivotDataStore = createPivotDataStore(
      pivotDashboardContext,
      pivotConfig,
    );

    cache.set(metricsViewName, pivotDataStore);

    return pivotDataStore;
  };
})();
