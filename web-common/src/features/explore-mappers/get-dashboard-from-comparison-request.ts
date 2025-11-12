import { getSortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { TransformerArgs } from "@rilldata/web-common/features/explore-mappers/types";
import { fillTimeRange } from "@rilldata/web-common/features/explore-mappers/utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1MetricsViewComparisonMeasureType,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";

export async function getDashboardFromComparisonRequest({
  req,
  dashboard,
  metricsView,
  explore,
  timeRangeSummary,
  executionTime,
}: TransformerArgs<V1MetricsViewComparisonRequest>) {
  if (req.where) dashboard.whereFilter = req.where;

  await fillTimeRange(
    explore,
    dashboard,
    req.timeRange,
    req.comparisonTimeRange,
    timeRangeSummary,
    executionTime,
  );

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  dashboard.visibleMeasures = req.measures?.map((m) => m.name ?? "") ?? [];

  // if the selected sort is a measure set it to leaderboardSortByMeasureName
  if (
    req.sort?.[0] &&
    (metricsView.measures?.findIndex((m) => m.name === req.sort?.[0]?.name) ??
      -1) >= 0
  ) {
    dashboard.leaderboardSortByMeasureName = req.sort[0].name ?? "";
    dashboard.sortDirection = req.sort[0].desc
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
    dashboard.dashboardSortType = getSortType(
      req.sort[0].sortType ??
        V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED,
    );
  }

  dashboard.selectedDimensionName = req.dimension?.name ?? "";
  dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;

  return dashboard;
}
