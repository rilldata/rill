import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

export enum ExploreStateURLParams {
  WebView = "view",

  LegacyProtoState = "state",

  TimeRange = "tr",
  TimeZone = "tz",
  ComparisonTimeRange = "compare_tr",
  TimeGrain = "grain",
  ComparisonDimension = "compare_dim",
  HighlightedTimeRange = "highlighted_tr",

  Filters = "f",

  VisibleMeasures = "measures",
  VisibleDimensions = "dims",
  ExpandedDimension = "expand_dim",
  SortBy = "sort_by",
  SortType = "sort_type",
  SortDirection = "sort_dir",

  LeaderboardMeasures = "leaderboard_measures",
  LeaderboardShowContextForAllMeasures = "lb_ctx",
  ExpandedMeasure = "measure",
  ChartType = "chart_type",
  Pin = "pin",

  PivotRows = "rows",
  PivotColumns = "cols",
  PivotTableMode = "table_mode",
  PivotRowLimit = "row_limit",

  GzippedParams = "gzipped_state",
  TemporaryFilters = "temp",
  TimeDimension = "td",

  IgnoreErrors = "ignore_errors",
}

export const ExploreStateKeyToURLParamMap: Partial<
  Record<keyof ExploreState, ExploreStateURLParams>
> = {
  activePage: ExploreStateURLParams.WebView,

  selectedTimezone: ExploreStateURLParams.TimeZone,
  selectedComparisonDimension: ExploreStateURLParams.ComparisonDimension,

  selectedDimensionName: ExploreStateURLParams.ExpandedDimension,
  leaderboardSortByMeasureName: ExploreStateURLParams.SortBy,
  dashboardSortType: ExploreStateURLParams.SortType,
  sortDirection: ExploreStateURLParams.SortDirection,
  leaderboardMeasureNames: ExploreStateURLParams.LeaderboardMeasures,
  leaderboardShowContextForAllMeasures:
    ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
};
