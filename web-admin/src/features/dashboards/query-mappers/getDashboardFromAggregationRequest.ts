import {
  convertExprToToplist,
  fillTimeRange,
} from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import type { QueryMapperArgs } from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import {
  createAndExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1Expression,
  V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";

export async function getDashboardFromAggregationRequest({
  queryClient,
  instanceId,
  req,
  dashboard,
  timeRangeSummary,
  executionTime,
  metricsView,
}: QueryMapperArgs<V1MetricsViewAggregationRequest>) {
  fillTimeRange(
    dashboard,
    req.timeRange,
    req.comparisonTimeRange,
    timeRangeSummary,
    executionTime,
  );

  if (req.where) dashboard.whereFilter = req.where;
  if (req.having?.cond?.exprs?.length && req.dimensions?.[0]?.name) {
    const dimension = req.dimensions[0].name;
    if (req.having.cond.exprs.length > 1 || exprHasComparison(req.having)) {
      const expr = await convertExprToToplist(
        queryClient,
        instanceId,
        dashboard.name,
        dimension,
        req.measures?.[0]?.name ?? "",
        req.timeRange,
        req.comparisonTimeRange,
        executionTime,
        req.where,
        req.having,
      );
      if (expr) {
        dashboard.whereFilter = mergeFilters(
          dashboard.whereFilter ?? createAndExpression([]),
          createAndExpression([expr]),
        );
      }
    } else {
      dashboard.dimensionThresholdFilters = [
        {
          name: dimension,
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

function exprHasComparison(expr: V1Expression) {
  let hasComparison = false;
  forEachIdentifier(expr, (e, ident) => {
    if (
      ident.endsWith(ComparisonDeltaAbsoluteSuffix) ||
      ident.endsWith(ComparisonDeltaRelativeSuffix)
    ) {
      hasComparison = true;
    }
  });
  return hasComparison;
}
