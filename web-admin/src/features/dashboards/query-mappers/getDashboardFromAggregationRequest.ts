import type { QueryMapperArgs } from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import { fillTimeRange } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  mapExprToMeasureFilter,
  measureHasSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import {
  createAndExpression,
  createSubQueryExpression,
  forEachIdentifier,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  getQueryServiceMetricsViewSchemaQueryKey,
  queryServiceMetricsViewSchema,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export async function getDashboardFromAggregationRequest({
  queryClient,
  instanceId,
  req,
  dashboard,
  timeRangeSummary,
  executionTime,
  metricsView,
  explore,
  annotations,
}: QueryMapperArgs<V1MetricsViewAggregationRequest>) {
  let loadedFromState = false;
  if (annotations["web_open_state"]) {
    await mergeDashboardFromUrlState(
      queryClient,
      instanceId,
      dashboard,
      metricsView,
      explore,
      annotations["web_open_state"],
    );
    loadedFromState = true;
  }

  fillTimeRange(
    dashboard,
    req.timeRange,
    req.comparisonTimeRange,
    timeRangeSummary,
    executionTime,
  );

  if (req.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      req.where,
    );
    dashboard.whereFilter = dimensionFilters;
    dashboard.dimensionThresholdFilters = dimensionThresholdFilters;
  }
  if (req.having?.cond?.exprs?.length && req.dimensions?.[0]?.name) {
    const dimension = req.dimensions[0].name;
    if (
      req.having.cond.exprs.length > 1 ||
      exprHasComparison(req.having) ||
      dashboard.dimensionThresholdFilters.length > 0
    ) {
      const extraFilter = createSubQueryExpression(
        dimension,
        getAllIdentifiers(req.having),
        req.having,
      );
      if (dashboard.whereFilter?.cond?.exprs?.length) {
        dashboard.whereFilter = createAndExpression([
          dashboard.whereFilter,
          extraFilter,
        ]);
      } else {
        dashboard.whereFilter = extraFilter;
      }
    } else {
      dashboard.dimensionThresholdFilters = [
        {
          name: dimension,
          filters:
            req.having.cond?.exprs
              ?.map(mapExprToMeasureFilter)
              .filter(Boolean) ?? [],
        },
      ];
    }
  }

  // everything after this can be loaded from the dashboard state if present
  if (loadedFromState) return dashboard;

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  dashboard.visibleMeasureKeys = new Set(
    req.measures
      ?.map((m) => m.name ?? "")
      .filter((m) => !measureHasSuffix(m)) ?? [],
  );
  dashboard.allMeasuresVisible =
    dashboard.visibleMeasureKeys.size === explore.measures.length;

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
      ident.endsWith(ComparisonDeltaRelativeSuffix) ||
      ident.endsWith(ComparisonPercentOfTotal)
    ) {
      hasComparison = true;
    }
  });
  return hasComparison;
}

async function mergeDashboardFromUrlState(
  queryClient: QueryClient,
  instanceId: string,
  dashboard: MetricsExplorerEntity,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  urlState: string,
) {
  const schemaResp = await queryClient.fetchQuery({
    queryKey: getQueryServiceMetricsViewSchemaQueryKey(
      instanceId,
      dashboard.name,
    ),
    queryFn: () => queryServiceMetricsViewSchema(instanceId, dashboard.name),
  });
  if (!schemaResp.schema) return;

  const parsedDashboard = getDashboardStateFromUrl(
    urlState,
    metricsViewSpec,
    exploreSpec,
    schemaResp.schema,
  );
  for (const k in parsedDashboard) {
    dashboard[k] = parsedDashboard[k];
  }
}
