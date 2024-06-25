import {
  PivotDataStoreConfig,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useStartEndTime } from "@rilldata/web-common/features/templates/kpi/selector";
import { TableProperties } from "@rilldata/web-common/features/templates/types";
import { Readable, derived } from "svelte/store";

export function getTableConfig(
  instanceId: string,
  tableProperties: TableProperties,
  pivotState: PivotState,
): Readable<PivotDataStoreConfig> {
  return derived(
    [
      useMetricsView(instanceId, tableProperties.metric_view),
      useStartEndTime(
        instanceId,
        tableProperties.metric_view,
        tableProperties.time_range,
      ),
    ],
    ([metricsView, timeRange]) => {
      const config: PivotDataStoreConfig = {
        measureNames: tableProperties.measures,
        rowDimensionNames: tableProperties.row_dimensions,
        colDimensionNames: tableProperties.col_dimensions,
        allMeasures: metricsView.data?.measures || [],
        allDimensions: metricsView.data?.dimensions || [],
        whereFilter: createAndExpression([]),
        pivot: pivotState,
        enableComparison: false,
        comparisonTime: undefined,
        time: {
          timeStart: timeRange?.data?.start?.toISOString() || undefined,
          timeEnd: timeRange?.data?.end?.toISOString() || undefined,
          timeZone: "UTC",
          timeDimension: "",
        },
      };

      return config;
    },
  );
}
