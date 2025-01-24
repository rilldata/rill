import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { canEnablePivotComparison } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDataStoreConfig,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { createQueryServiceMetricsViewTimeRange } from "@rilldata/web-common/runtime-client";
import { type Readable, derived } from "svelte/store";

export function useComparisonStartEndTime(
  instanceId: string,
  metricsViewName: string,
  timeRange: string,
  comparisonRange: string | undefined,
) {
  const allTimeRangeQuery = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
  );
  return derived(allTimeRangeQuery, (allTimeRange) => {
    const maxTime = allTimeRange?.data?.timeRangeSummary?.max;
    const maxTimeDate = new Date(maxTime ?? 0);
    const { startTime } = isoDurationToTimeRange(timeRange, maxTimeDate);

    let comparisonStartTime: Date | undefined = undefined;
    let comparisonEndTime: Date | undefined = undefined;

    if (comparisonRange) {
      ({ startTime: comparisonStartTime, endTime: comparisonEndTime } =
        isoDurationToTimeRange(comparisonRange, startTime));
    }
    return { start: comparisonStartTime, end: comparisonEndTime };
  });
}

export function useStartEndTime(
  instanceId: string,
  metricsViewName: string,
  timeRange: string,
) {
  return createQueryServiceMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
    {
      query: {
        select: (data) => {
          const maxTime = new Date(data?.timeRangeSummary?.max ?? 0);
          const { startTime, endTime } = isoDurationToTimeRange(
            timeRange,
            maxTime,
          );

          return { start: startTime, end: endTime };
        },
      },
    },
  );
}

export function getTableConfig(
  ctx: StateManagers,
  tableSpec: TableSpec,
  pivotState: PivotState,
): Readable<PivotDataStoreConfig> {
  const { metrics_view, time_range, comparison_range, dimension_filters } =
    tableSpec;
  const {
    canvasEntity: {
      createTimeAndFilterStore,
      spec: { getMetricsViewFromName },
    },
  } = ctx;

  const timeAndFilterStore = createTimeAndFilterStore(metrics_view, {
    componentTimeRange: time_range,
    componentComparisonRange: comparison_range,
    componentFilter: dimension_filters,
  });

  return derived(
    [getMetricsViewFromName(metrics_view), timeAndFilterStore],
    ([metricsView, { timeRange, comparisonRange, where }]) => {
      const enableComparison = canEnablePivotComparison(
        pivotState,
        comparisonRange.start,
      );

      const config: PivotDataStoreConfig = {
        measureNames: tableSpec.measures.flatMap((name) => {
          const group = [name];
          if (enableComparison) {
            group.push(
              `${name}${COMPARISON_DELTA}`,
              `${name}${COMPARISON_PERCENT}`,
            );
          }
          return group;
        }),
        rowDimensionNames: tableSpec.row_dimensions || [],
        colDimensionNames: tableSpec.col_dimensions || [],
        allMeasures: metricsView?.measures || [],
        allDimensions: metricsView?.dimensions || [],
        whereFilter: where ?? createAndExpression([]),
        searchText: "",
        pivot: pivotState,
        enableComparison,
        comparisonTime: {
          start: comparisonRange?.start,
          end: comparisonRange?.end,
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
      const measures = tableSpec.measures;
      const rowDimensions = tableSpec.row_dimensions || [];
      const colDimensions = tableSpec.col_dimensions || [];

      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }
      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure(s) ${invalidMeasures} selected for the table`,
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
          error: `Invalid dimension(s) ${invalidDimensions} selected for the table`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
