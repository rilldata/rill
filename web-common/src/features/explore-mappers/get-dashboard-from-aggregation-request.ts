import { splitDimensionsAndMeasuresAsRowsAndColumns } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  mapExprToMeasureFilter,
  measureHasSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  COMPARISON_VALUE,
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  createSubQueryExpression,
  forEachIdentifier,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import type { TransformerArgs } from "@rilldata/web-common/features/explore-mappers/types";
import {
  convertQueryFilterToToplistQuery,
  fillTimeRange,
} from "@rilldata/web-common/features/explore-mappers/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import type { SortingState } from "@tanstack/svelte-table";

export async function getDashboardFromAggregationRequest({
  queryClient,
  instanceId,
  req,
  dashboard,
  timeRangeSummary,
  executionTime,
  metricsView,
  explore,
  exploreProtoState,
  ignoreFilters,
  forceOpenPivot,
}: TransformerArgs<V1MetricsViewAggregationRequest>) {
  let loadedFromState = false;
  if (exploreProtoState) {
    await mergeDashboardFromUrlState(
      queryClient,
      instanceId,
      dashboard,
      metricsView,
      explore,
      exploreProtoState,
    );
    loadedFromState = true;
  }

  await fillTimeRange(
    explore,
    dashboard,
    req.timeRange,
    req.comparisonTimeRange,
    timeRangeSummary,
    executionTime,
  );

  const shouldParseWhereFilter = Boolean(!ignoreFilters && req.where);
  if (shouldParseWhereFilter) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      req.where,
    );
    dashboard.whereFilter = dimensionFilters;
    dashboard.dimensionThresholdFilters = dimensionThresholdFilters;
  }

  const shouldParseHavingFilter = Boolean(
    !ignoreFilters &&
      req.having?.cond?.exprs?.length &&
      req.dimensions?.[0]?.name,
  );
  if (shouldParseHavingFilter) {
    const dimension = req.dimensions![0].name!;
    if (exprHasComparison(req.having!)) {
      // We do not support comparison based dimension threshold filter in dashboards right now.
      // So convert it to a toplist and add `in` filter.
      const expr = await convertQueryFilterToToplistQuery(
        instanceId,
        explore.metricsView ?? "",
        req,
        dimension,
      );
      dashboard.whereFilter =
        mergeFilters(
          dashboard.whereFilter ?? createAndExpression([]),
          createAndExpression([expr]),
        ) ?? createAndExpression([]);
    } else if (
      req.having!.cond!.exprs!.length > 1 ||
      dashboard.dimensionThresholdFilters.length > 0
    ) {
      // If there are dimension threshold and having filter we just add a subquery in where filter.
      // This will be marked as "advanced filter" that is not editable.
      // TODO: find a way to merge having filter into dimension threshold
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
            req.having?.cond?.exprs
              ?.map(mapExprToMeasureFilter)
              .filter((f): f is NonNullable<typeof f> => f != null) ?? [],
        },
      ];
    }
  }

  // everything after this can be loaded from the dashboard state if present
  if (loadedFromState) return dashboard;

  if (req.timeRange?.timeZone) {
    dashboard.selectedTimezone = req.timeRange?.timeZone || "UTC";
  }

  if (forceOpenPivot) {
    dashboard.activePage = DashboardState_ActivePage.PIVOT;
    dashboard.pivot = getPivotStateFromRequest(req);
    return dashboard;
  }

  if (req.measures?.length) {
    dashboard.visibleMeasures = req.measures
      .map((m) => m.name ?? "")
      .filter((m) => !measureHasSuffix(m));
    dashboard.allMeasuresVisible =
      dashboard.visibleMeasures.length === explore.measures?.length;
  }

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
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  urlState: string,
) {
  if (!exploreSpec.metricsView) return;

  const parsedDashboard = getDashboardStateFromUrl(
    urlState,
    metricsViewSpec,
    exploreSpec,
  );
  for (const k in parsedDashboard) {
    exploreState[k] = parsedDashboard[k];
  }
}

function getPivotStateFromRequest(
  req: V1MetricsViewAggregationRequest,
): PivotState {
  const { rows, dimensionColumns, measureColumns } =
    splitDimensionsAndMeasuresAsRowsAndColumns(req);

  const mapDimension = (
    dim: V1MetricsViewAggregationDimension,
  ): PivotChipData => {
    if (dim.timeGrain && dim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
      return {
        id: dim.timeGrain,
        title: TIME_GRAIN[dim.timeGrain]?.label,
        type: PivotChipType.Time,
      };
    }

    return {
      id: dim.name!,
      title: dim.alias ?? dim.name ?? "",
      type: PivotChipType.Dimension,
    };
  };
  const mapMeasure = (mes: V1MetricsViewAggregationMeasure): PivotChipData => {
    return {
      id: mes.name!,
      title: mes.name!,
      type: PivotChipType.Measure,
    };
  };

  const rowChips: PivotChipData[] = rows.map(mapDimension);

  const colChips: PivotChipData[] = [
    ...dimensionColumns.map(mapDimension),
    ...measureColumns.map(mapMeasure),
  ];

  const isFlat = !req.pivotOn?.length;

  const tableMode = isFlat ? "flat" : "nest";

  const sorting: SortingState =
    req.sort?.map((s) => ({
      id: convertComparisonMeasureToPivotMeasures(s.name!),
      desc: !!s.desc,
    })) ?? [];

  return {
    rows: rowChips,
    columns: colChips,
    sorting,
    expanded: {},
    columnPage: 1,
    rowPage: 1,
    enableComparison: true, // This is not really used. So setting it true, we should remove it.
    activeCell: null,
    tableMode,
  };
}

// We use a different suffix in pivot vs rest of the app. So map them correctly to not break pivot.
// TODO: Unify the suffix to avoid this confusion
function convertComparisonMeasureToPivotMeasures(measure: string) {
  switch (true) {
    case measure.endsWith(ComparisonDeltaPreviousSuffix):
      return measure.replace(ComparisonDeltaPreviousSuffix, COMPARISON_VALUE);

    case measure.endsWith(ComparisonDeltaAbsoluteSuffix):
      return measure.replace(ComparisonDeltaAbsoluteSuffix, COMPARISON_DELTA);

    case measure.endsWith(ComparisonDeltaRelativeSuffix):
      return measure.replace(ComparisonDeltaRelativeSuffix, COMPARISON_PERCENT);

    case measure.endsWith(ComparisonPercentOfTotal):
      return measure.replace(ComparisonPercentOfTotal, "");
  }

  return measure;
}
