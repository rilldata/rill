import {
  DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_PivotTotalsRowPosition,
  DashboardTimeRange,
  PivotElement,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import type { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb.ts";
import {
  base64ToProto,
  correctComparisonTimeRange,
  fromExpressionProto,
  fromPivotConditionalFormattingProto,
} from "@rilldata/web-common/features/dashboards/proto-state/fromProto.ts";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter.ts";
import {
  createAndExpression,
  createSubQueryExpression,
  getAllIdentifiers,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import {
  FromProtoPivotTableModeMap,
  FromProtoTimeGrainMap,
} from "@rilldata/web-common/features/dashboards/proto-state/enum-maps.ts";
import {
  FromActivePageMap,
  ToURLParamSortTypeMap,
  ToURLParamTimeDimensionMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";
import {
  FromLegacySortTypeMap,
  mapLegacyChartType,
} from "@rilldata/web-common/features/dashboards/url-state/legacyMappers.ts";
import { toEphemeralMeasuresParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param.ts";
import { toPivotFormattingParam } from "@rilldata/web-common/features/dashboards/pivot/pivot-formatting-param.ts";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import { V1TimeGrainToToken } from "@rilldata/web-common/lib/time/new-grains.ts";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";
import {
  convertAbsoluteTimeRangeToRillTime,
  convertAbsoluteTimeRangeToUrlParam,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";

/**
 * Converts the legacy base64 encoded proto dashboard state to explore url params.
 * This does not need the explore or metrics view specs:
 * the params are validated against them when the explore is built from the url in DashboardStateSync.
 *
 * Fields with no url param are dropped:
 * leaderboard context column, tdd pin, pivot expanded state, pivot column page and pivot comparison.
 */
export function convertProtoToUrlParams(protoState: string) {
  const dashboard = DashboardState.fromBinary(base64ToProto(protoState));
  const urlParams = new URLSearchParams();

  // UNSPECIFIED means a partial state was stored.
  // A full state also unsets fields that it does not have, so the explore's defaults do not fill them in.
  const isFullState =
    dashboard.activePage !== DashboardState_ActivePage.UNSPECIFIED;
  const view = ToURLParamViewMap[FromActivePageMap[dashboard.activePage]];
  if (isFullState && view) {
    urlParams.set(ExploreStateURLParams.WebView, view);
  }

  if (dashboard.ephemeralMeasures.length) {
    urlParams.set(
      ExploreStateURLParams.EphemeralMeasures,
      toEphemeralMeasuresParam(
        dashboard.ephemeralMeasures.map((def) => ({
          name: def.name,
          displayName: def.displayName,
          expression: def.expression,
          ...(def.formatPreset ? { formatPreset: def.formatPreset } : {}),
        })),
      ),
    );
  }

  copyParamsToTarget(convertProtoFiltersToUrlParams(dashboard), urlParams);
  copyParamsToTarget(
    convertProtoTimeFieldsToUrlParams(dashboard, isFullState),
    urlParams,
  );
  if (dashboard.activePage === DashboardState_ActivePage.PIVOT) {
    copyParamsToTarget(
      convertProtoPivotFieldsToUrlParams(dashboard),
      urlParams,
    );
  } else {
    // Explore and pivot share the sort params, so only one of them can be set.
    copyParamsToTarget(
      convertProtoExploreFieldsToUrlParams(dashboard, isFullState),
      urlParams,
    );
  }
  copyParamsToTarget(
    convertProtoTimeDimensionFieldsToUrlParams(dashboard),
    urlParams,
  );

  return urlParams;
}

function convertProtoFiltersToUrlParams(dashboard: DashboardState) {
  const urlParams = new URLSearchParams();

  let expr: V1Expression | undefined = undefined;

  if (dashboard.filters) {
    // backwards compatibility for our older filter format
    expr = convertFilterToExpression(dashboard.filters);
  } else if (dashboard.where) {
    expr = fromExpressionProto(dashboard.where);
  }

  // Explore state keeps measure filters collapsed into the where filter as subqueries.
  // Older protos stored them separately in `having`, so merge those back in.
  if (dashboard.having.length) {
    expr = maybeWrapAndExpression(expr) ?? createAndExpression([]);

    dashboard.having.forEach((h) => {
      if (!h.filter) return;
      const havingExpr = fromExpressionProto(h.filter);
      expr?.cond?.exprs?.push(
        createSubQueryExpression(
          h.name,
          getAllIdentifiers(havingExpr),
          havingExpr,
        ),
      );
    });
  }

  const param = convertExpressionToFilterParam(
    expr,
    dashboard.dimensionsWithInlistFilter ?? [],
  );
  if (param) urlParams.set(ExploreStateURLParams.Filters, param);

  return urlParams;
}

function convertProtoTimeFieldsToUrlParams(
  dashboard: DashboardState,
  isFullState: boolean,
) {
  const urlParams = new URLSearchParams();

  if (dashboard.timeRange) {
    const timeRangeName = convertProtoTimeRangeToUrlParam(
      dashboard.timeRange,
      // Rill time reads an absolute time range as wall-clock times in the dashboard's zone.
      dashboard.selectedTimezone || "UTC",
    );
    if (timeRangeName) {
      urlParams.set(ExploreStateURLParams.TimeRange, timeRangeName);
    }

    // The grain only applies along with its time range; without it the explore keeps its adaptive grain.
    if (dashboard.timeGrain) {
      const grain =
        V1TimeGrainToToken[FromProtoTimeGrainMap[dashboard.timeGrain]];
      if (grain) urlParams.set(ExploreStateURLParams.TimeGrain, grain);
    }
  }

  // `showTimeComparison` is always written to the proto, so it is only read along with a comparison range,
  // or when it turns the comparison on.
  if (dashboard.compareTimeRange) {
    // backwards compatibility
    const compareTimeRangeName = dashboard.showTimeComparison
      ? correctComparisonTimeRange(
          convertProtoTimeRangeToUrlParam(dashboard.compareTimeRange) ?? "",
        )
      : "";
    urlParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      compareTimeRangeName,
    );
  } else if (dashboard.showTimeComparison) {
    // The explore used to derive the comparison range from its time range config,
    // which needs the specs and the full time range; the url has no param for that,
    // so this falls back to the previous period, the default when the explore configures no comparison.
    // Older encoders always wrote the comparison range along with the flag, so this is not expected in practice.
    urlParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      TimeComparisonOption.CONTIGUOUS,
    );
  }

  if (dashboard.comparisonDimension) {
    urlParams.set(
      ExploreStateURLParams.ComparisonDimension,
      dashboard.comparisonDimension,
    );
  } else if (isFullState) {
    urlParams.set(ExploreStateURLParams.ComparisonDimension, "");
  }

  if (dashboard.scrubRange) {
    const scrubRangeName = convertProtoTimeRangeToUrlParam(
      dashboard.scrubRange,
    );
    if (scrubRangeName) {
      urlParams.set(ExploreStateURLParams.HighlightedTimeRange, scrubRangeName);
    }
  }

  if (dashboard.selectedTimezone !== undefined) {
    urlParams.set(ExploreStateURLParams.TimeZone, dashboard.selectedTimezone);
  } else if (dashboard.timeRange?.timeStart && dashboard.timeRange.timeEnd) {
    // The absolute time range was written in UTC, so the explore's default zone must not apply.
    urlParams.set(ExploreStateURLParams.TimeZone, "UTC");
  }

  return urlParams;
}

/**
 * @param timeZone When set, an absolute range is written as a rill time range of wall-clock times in the zone,
 * as the primary time range param needs.
 * Otherwise it is written as UTC instants, as the comparison and highlighted time range params need.
 */
function convertProtoTimeRangeToUrlParam(
  timeRange: DashboardTimeRange,
  timeZone?: string,
) {
  const rangeName = timeRange.name;

  // Check custom first since TimeRangePreset has CUSTOM as well
  if (timeRange.timeStart && timeRange.timeEnd) {
    // The legacy encoder stored milliseconds in `seconds`, so `Timestamp.toDate` cannot be used.
    const start = new Date(Number(timeRange.timeStart.seconds));
    const end = new Date(Number(timeRange.timeEnd.seconds));
    return timeZone
      ? convertAbsoluteTimeRangeToRillTime(start, end, timeZone)
      : convertAbsoluteTimeRangeToUrlParam(start, end);
  }

  // backwards compatibility
  if (rangeName && rangeName in TimeRangePreset) {
    return TimeRangePreset[rangeName] as string;
  }

  return rangeName;
}

function convertProtoExploreFieldsToUrlParams(
  dashboard: DashboardState,
  isFullState: boolean,
) {
  const urlParams = new URLSearchParams();

  if (dashboard.allMeasuresVisible) {
    urlParams.set(ExploreStateURLParams.VisibleMeasures, "*");
  } else if (dashboard.visibleMeasures.length) {
    urlParams.set(
      ExploreStateURLParams.VisibleMeasures,
      dashboard.visibleMeasures.join(","),
    );
  }

  if (dashboard.allDimensionsVisible) {
    urlParams.set(ExploreStateURLParams.VisibleDimensions, "*");
  } else if (dashboard.visibleDimensions.length) {
    urlParams.set(
      ExploreStateURLParams.VisibleDimensions,
      dashboard.visibleDimensions.join(","),
    );
  }

  if (dashboard.leaderboardMeasure) {
    urlParams.set(ExploreStateURLParams.SortBy, dashboard.leaderboardMeasure);
  }
  if (dashboard.leaderboardSortDirection) {
    urlParams.set(
      ExploreStateURLParams.SortDirection,
      dashboard.leaderboardSortDirection ===
        DashboardState_LeaderboardSortDirection.ASCENDING
        ? "ASC"
        : "DESC",
    );
  }
  if (dashboard.leaderboardSortType) {
    const sortType =
      ToURLParamSortTypeMap[
        FromLegacySortTypeMap[dashboard.leaderboardSortType]
      ];
    if (sortType) urlParams.set(ExploreStateURLParams.SortType, sortType);
  }

  if (dashboard.leaderboardMeasures.length) {
    urlParams.set(
      ExploreStateURLParams.LeaderboardMeasures,
      dashboard.leaderboardMeasures.join(","),
    );
  }
  if (dashboard.leaderboardShowContextForAllMeasures !== undefined) {
    urlParams.set(
      ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
      String(dashboard.leaderboardShowContextForAllMeasures),
    );
  }

  if (dashboard.selectedDimension) {
    urlParams.set(
      ExploreStateURLParams.ExpandedDimension,
      dashboard.selectedDimension,
    );
  } else if (isFullState) {
    urlParams.set(ExploreStateURLParams.ExpandedDimension, "");
  }

  return urlParams;
}

function convertProtoTimeDimensionFieldsToUrlParams(dashboard: DashboardState) {
  const urlParams = new URLSearchParams();

  if (dashboard.expandedMeasure) {
    urlParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      dashboard.expandedMeasure,
    );
  } else if (
    dashboard.activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
  ) {
    urlParams.set(ExploreStateURLParams.ExpandedMeasure, "");
  }

  if (dashboard.chartType) {
    urlParams.set(
      ExploreStateURLParams.ChartType,
      mapLegacyChartType(dashboard.chartType),
    );
  }

  return urlParams;
}

function convertProtoPivotFieldsToUrlParams(dashboard: DashboardState) {
  const urlParams = new URLSearchParams();

  const mapTimeDimension = (grain: TimeGrain) =>
    ToURLParamTimeDimensionMap[FromProtoTimeGrainMap[grain]] ?? "";
  const mapAllDimension = (dimension: PivotElement) =>
    dimension.element.case === "pivotTimeDimension"
      ? mapTimeDimension(dimension.element.value)
      : (dimension.element.value as string);

  let rows: string[] = [];
  let cols: string[] = [];
  if (
    dashboard.pivotRowAllDimensions.length ||
    dashboard.pivotColumnAllDimensions.length
  ) {
    rows = dashboard.pivotRowAllDimensions.map(mapAllDimension);
    cols = dashboard.pivotColumnAllDimensions.map(mapAllDimension);
  } else {
    // backwards compatibility for state
    rows = [
      ...dashboard.pivotRowTimeDimensions.map(mapTimeDimension),
      ...dashboard.pivotRowDimensions,
    ];
    cols = [
      ...dashboard.pivotColumnTimeDimensions.map(mapTimeDimension),
      ...dashboard.pivotColumnDimensions,
    ];
  }
  cols.push(...dashboard.pivotColumnMeasures);

  urlParams.set(
    ExploreStateURLParams.PivotRows,
    rows.filter(Boolean).join(","),
  );
  urlParams.set(
    ExploreStateURLParams.PivotColumns,
    cols.filter(Boolean).join(","),
  );

  const sort = dashboard.pivotSort[0];
  if (sort) {
    urlParams.set(
      ExploreStateURLParams.SortBy,
      convertPivotSortIdToUrlParam(sort.id),
    );
    urlParams.set(
      ExploreStateURLParams.SortDirection,
      sort.desc ? "DESC" : "ASC",
    );
  }

  if (dashboard.pivotTableMode) {
    urlParams.set(
      ExploreStateURLParams.PivotTableMode,
      FromProtoPivotTableModeMap[dashboard.pivotTableMode],
    );
  }
  if (dashboard.pivotRowLimit !== undefined) {
    urlParams.set(
      ExploreStateURLParams.PivotRowLimit,
      String(dashboard.pivotRowLimit),
    );
  }
  if (dashboard.pivotShowTotalsColumn === false) {
    urlParams.set(ExploreStateURLParams.PivotShowTotalsColumn, "false");
  }
  if (dashboard.pivotShowTotalsRow === false) {
    urlParams.set(ExploreStateURLParams.PivotShowTotalsRow, "false");
  }
  if (
    dashboard.pivotTotalsRowPosition ===
    DashboardState_PivotTotalsRowPosition.BOTTOM
  ) {
    urlParams.set(ExploreStateURLParams.PivotTotalsRowPosition, "bottom");
  }
  if (dashboard.pivotConditionalFormatting.length) {
    urlParams.set(
      ExploreStateURLParams.PivotFormatting,
      toPivotFormattingParam(
        fromPivotConditionalFormattingProto(
          dashboard.pivotConditionalFormatting,
        ),
      ),
    );
  }

  return urlParams;
}

// The pivot sorts a time column either by its grain or by `{timeDimension}_rill_{grain}`.
function convertPivotSortIdToUrlParam(sortId: string) {
  if (sortId in ToURLParamTimeDimensionMap) {
    return ToURLParamTimeDimensionMap[sortId] as string;
  }
  const grain = sortId.split("_rill_")[1];
  if (grain && grain in ToURLParamTimeDimensionMap) {
    return ToURLParamTimeDimensionMap[grain] as string;
  }
  return sortId;
}
