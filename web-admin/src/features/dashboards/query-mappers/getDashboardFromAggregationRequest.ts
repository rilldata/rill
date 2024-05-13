import {
  convertExprToToplist,
  getSelectedTimeRange,
} from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import type { QueryMapperArgs } from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1MetricsViewAggregationRequest } from "@rilldata/web-common/runtime-client";

export async function getDashboardFromAggregationRequest({
  queryClient,
  instanceId,
  req,
  dashboard,
  timeRangeSummary,
  executionTime,
  metricsView,
}: QueryMapperArgs<V1MetricsViewAggregationRequest>) {
  if (req.timeRange) {
    dashboard.selectedTimeRange = getSelectedTimeRange(
      req.timeRange,
      timeRangeSummary,
      req.timeRange.isoDuration ?? "",
      executionTime,
    );
  }

  if (req.where) dashboard.whereFilter = req.where;
  if (req.having?.cond?.exprs?.length && req.dimensions?.[0]?.name) {
    if (req.having.cond.exprs.length > 1) {
      const expr = await convertExprToToplist(
        queryClient,
        instanceId,
        dashboard.name,
        req.dimensions[0].name,
        req.measures?.[0]?.name ?? "",
        dashboard.selectedTimeRange,
        req.where,
        req.having.cond.exprs ?? [],
      );
      if (expr) {
        dashboard.whereFilter ??= createAndExpression([]);
        if (dashboard.whereFilter.cond?.exprs)
          dashboard.whereFilter.cond.exprs.push(expr);
      }
    } else {
      dashboard.dimensionThresholdFilters = [
        {
          name: req.dimensions[0].name,
          filter: createAndExpression([req.having.cond?.exprs?.[0]]),
        },
      ];
    }
  }

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  dashboard.visibleMeasureKeys = new Set(
    req.measures?.map((m) => m.name ?? "") ?? [],
  );

  // if the selected sort is a measure set it to leaderboardMeasureName
  if (
    req.sort?.[0] &&
    (metricsView.measures?.findIndex((m) => m.name === req.sort?.[0]?.name) ??
      -1) >= 0
  ) {
    dashboard.leaderboardMeasureName = req.sort[0].name ?? "";
    dashboard.sortDirection = req.sort[0].desc
      ? SortDirection.DESCENDING
      : SortDirection.ASCENDING;
    dashboard.dashboardSortType = SortType.VALUE;
  }

  if (req.dimensions?.length) {
    dashboard.selectedDimensionName = req.dimensions[0].name;
    dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  } else {
    dashboard.tdd = {
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: req.measures?.[0]?.name ?? "",
      pinIndex: -1,
    };
    dashboard.activePage = DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
  }

  return dashboard;
}
