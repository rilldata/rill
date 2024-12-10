import { useStartEndTime } from "@rilldata/web-common/features/canvas/components/kpi/selector";
import { canEnablePivotComparison } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotDataStoreConfig,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  useMetricsViewTimeRange,
  useMetricsViewValidSpec,
} from "@rilldata/web-common/features/dashboards/selectors";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TableProperties } from "@rilldata/web-common/features/templates/types";
import {
  validateDimensions,
  validateMeasures,
  validateMetricsView,
} from "@rilldata/web-common/features/templates/utils";
import { isoDurationToTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
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

export function getTableConfig(
  instanceId: string,
  tableProperties: TableProperties,
  pivotState: PivotState,
): Readable<PivotDataStoreConfig> {
  const { metrics_view, time_range, comparison_range } = tableProperties;
  return derived(
    [
      useMetricsViewValidSpec(instanceId, tableProperties.metrics_view),
      useStartEndTime(instanceId, metrics_view, time_range.toUpperCase()),
      useComparisonStartEndTime(
        instanceId,
        metrics_view,
        time_range.toUpperCase(),
        comparison_range,
      ),
    ],
    ([metricsView, timeRange, comparisonRange]) => {
      const enableComparison = canEnablePivotComparison(
        pivotState,
        comparisonRange.start,
      );

      const config: PivotDataStoreConfig = {
        measureNames: tableProperties.measures.flatMap((name) => {
          const group = [name];
          if (enableComparison) {
            group.push(
              `${name}${COMPARISON_DELTA}`,
              `${name}${COMPARISON_PERCENT}`,
            );
          }
          return group;
        }),
        rowDimensionNames: tableProperties.row_dimensions || [],
        colDimensionNames: tableProperties.col_dimensions || [],
        allMeasures: metricsView.data?.measures || [],
        allDimensions: metricsView.data?.dimensions || [],
        whereFilter: createAndExpression([]),
        searchText: "",
        pivot: pivotState,
        enableComparison,
        comparisonTime: {
          start: comparisonRange?.start?.toISOString() || undefined,
          end: comparisonRange?.end?.toISOString() || undefined,
        },
        time: {
          timeStart: timeRange?.data?.start?.toISOString() || undefined,
          timeEnd: timeRange?.data?.end?.toISOString() || undefined,
          timeZone: "UTC",
          timeDimension: metricsView?.data?.timeDimension || "",
        },
      };

      return config;
    },
  );
}

export function hasValidTableSchema(
  instanceId: string,
  tableProperties: TableProperties,
) {
  return derived(
    [useMetricsViewValidSpec(instanceId, tableProperties.metrics_view)],
    ([metricsView]) => {
      const measures = tableProperties.measures;
      const rowDimensions = tableProperties.row_dimensions || [];
      const colDimensions = tableProperties.col_dimensions || [];

      const validateMetricsViewRes = validateMetricsView(metricsView);

      if (!validateMetricsViewRes.isValid) {
        return {
          isValid: false,
          error: validateMetricsViewRes.error,
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
