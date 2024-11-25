import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  URLStateDefaultSortDirection,
  URLStateDefaultTDDChartType,
  URLStateDefaultTimeRange,
  URLStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  FromActivePageMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
  ToURLParamTimeGrainMapMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  arrayOrderedEquals,
  arrayUnorderedEquals,
} from "@rilldata/web-common/lib/arrayUtils";
import { inferCompareTimeRange } from "@rilldata/web-common/lib/time/comparisons";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export function convertMetricsEntityToURLSearchParams(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (!exploreState) return searchParams;

  const currentView = FromActivePageMap[exploreState.activePage];
  if (
    (preset.view !== undefined && preset.view !== currentView) ||
    (preset.view === undefined &&
      currentView !== V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW)
  ) {
    searchParams.set("vw", ToURLParamViewMap[currentView] as string);
  }

  const expr = mergeMeasureFilters(exploreState);
  if (expr && expr?.cond?.exprs?.length) {
    searchParams.set("f", convertExpressionToFilterParam(expr));
  }

  mergeSearchParams(
    toTimeRangesUrl(exploreState, exploreSpec, preset),
    searchParams,
  );

  mergeSearchParams(
    toOverviewUrl(exploreState, exploreSpec, preset),
    searchParams,
  );

  mergeSearchParams(
    toTimeDimensionUrlParams(exploreState, preset),
    searchParams,
  );

  mergeSearchParams(toPivotUrlParams(exploreState, preset), searchParams);

  return searchParams;
}

function toTimeRangesUrl(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (
    (preset.timeRange !== undefined &&
      exploreState.selectedTimeRange !== undefined &&
      exploreState.selectedTimeRange.name !== preset.timeRange) ||
    (preset.timeRange === undefined &&
      exploreState.selectedTimeRange?.name !== URLStateDefaultTimeRange)
  ) {
    searchParams.set("tr", exploreState.selectedTimeRange?.name ?? "");
  }

  const mappedTimeGrain =
    ToURLParamTimeGrainMapMap[exploreState.selectedTimeRange?.interval ?? ""] ??
    "";
  if (
    // if preset has a time grain, only set if selected is not the same
    (preset.timeGrain !== undefined && mappedTimeGrain !== preset.timeGrain) ||
    // else if there is no default then set if there was a selected time grain
    (preset.timeGrain === undefined && !!mappedTimeGrain)
  ) {
    searchParams.set("tg", mappedTimeGrain);
  }

  if (
    // if preset has timezone, only set if selected is not the same
    (preset.timezone !== undefined &&
      exploreState.selectedTimezone !== preset.timezone) ||
    // else if the timezone is not the default then set the param
    (preset.timezone === undefined &&
      exploreState.selectedTimezone !== URLStateDefaultTimezone)
  ) {
    searchParams.set("tz", exploreState.selectedTimezone);
  }

  if (exploreState.showTimeComparison) {
    if (
      (preset.compareTimeRange !== undefined &&
        exploreState.selectedComparisonTimeRange !== undefined &&
        exploreState.selectedComparisonTimeRange.name !==
          preset.compareTimeRange) ||
      preset.compareTimeRange === undefined
    ) {
      searchParams.set(
        "ctr",
        exploreState.selectedComparisonTimeRange?.name ?? "",
      );
    } else if (
      !exploreState.selectedComparisonTimeRange?.name &&
      exploreState.selectedTimeRange?.name
    ) {
      // we infer compare time range if the user has not explicitly selected one but has enabled comparison
      const inferredCompareTimeRange = inferCompareTimeRange(
        exploreSpec.timeRanges,
        exploreState.selectedTimeRange.name,
      );
      if (inferredCompareTimeRange)
        searchParams.set("ctr", inferredCompareTimeRange);
    }
  }

  if (
    // if preset has a compare dimension, only set if selected is not the same
    (preset.comparisonDimension !== undefined &&
      exploreState.selectedComparisonDimension !==
        preset.comparisonDimension) ||
    // else if there is no default then set if there was a selected compare dimension
    (preset.comparisonDimension === undefined &&
      !!exploreState.selectedComparisonDimension)
  ) {
    searchParams.set("cd", exploreState.selectedComparisonDimension ?? "");
  }

  return searchParams;
}

function toOverviewUrl(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  const visibleMeasuresParam = toVisibleMeasuresUrlParam(
    exploreState,
    exploreSpec,
    preset,
  );
  if (visibleMeasuresParam) {
    searchParams.set("o.m", visibleMeasuresParam);
  }

  const visibleDimensionsParam = toVisibleDimensionsUrlParam(
    exploreState,
    exploreSpec,
    preset,
  );
  if (visibleDimensionsParam) {
    searchParams.set("o.d", visibleDimensionsParam);
  }

  if (
    (preset.overviewExpandedDimension !== undefined &&
      exploreState.selectedDimensionName !==
        preset.overviewExpandedDimension) ||
    (preset.overviewExpandedDimension === undefined &&
      exploreState.selectedDimensionName)
  ) {
    searchParams.set("o.ed", exploreState.selectedDimensionName ?? "");
  }

  const defaultLeaderboardMeasure =
    preset.measures?.[0] ?? exploreSpec.measures?.[0];
  if (
    // if sort by is defined in preset then only set param if selected is not the same.
    (preset.overviewSortBy &&
      exploreState.leaderboardMeasureName !== preset.overviewSortBy) ||
    // else the default is the 1st measure in preset or exploreSpec, so check that next
    (!preset.overviewSortBy &&
      exploreState.leaderboardMeasureName !== defaultLeaderboardMeasure)
  ) {
    searchParams.set("o.sb", exploreState.leaderboardMeasureName);
  }

  const sortAsc = exploreState.sortDirection === SortDirection.ASCENDING;
  if (
    // if preset has a sort direction then only set if not the same
    (preset.overviewSortAsc !== undefined &&
      preset.overviewSortAsc !== sortAsc) ||
    // else if the direction is not the default then set the param
    (preset.overviewSortAsc === undefined &&
      exploreState.sortDirection !== URLStateDefaultSortDirection)
  ) {
    searchParams.set("o.sd", sortAsc ? "ASC" : "DESC");
  }

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleMeasureKeys) return undefined;

  const measures = [...exploreState.visibleMeasureKeys];
  const presetMeasures = preset.measures ?? exploreSpec.measures ?? [];
  if (arrayUnorderedEquals(measures, presetMeasures)) {
    return undefined;
  }
  if (exploreState.allMeasuresVisible) {
    return "*";
  }
  return measures.join(",");
}

function toVisibleDimensionsUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleDimensionKeys) return undefined;

  const dimensions = [...exploreState.visibleDimensionKeys];
  const presetDimensions = preset.dimensions ?? exploreSpec.dimensions ?? [];
  if (arrayUnorderedEquals(dimensions, presetDimensions)) {
    return undefined;
  }
  if (exploreState.allDimensionsVisible) {
    return "*";
  }
  return dimensions.join(",");
}

function toTimeDimensionUrlParams(
  exploreState: MetricsExplorerEntity,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (
    (preset.timeDimensionMeasure !== undefined &&
      exploreState.tdd.expandedMeasureName !== preset.timeDimensionMeasure) ||
    (preset.timeDimensionMeasure === undefined &&
      exploreState.tdd.expandedMeasureName)
  ) {
    searchParams.set("tdd.m", exploreState.tdd.expandedMeasureName ?? "");
  }

  if (
    (preset.timeDimensionChartType !== undefined &&
      ToURLParamTDDChartMap[exploreState.tdd.chartType] !==
        preset.timeDimensionChartType) ||
    (preset.timeDimensionChartType === undefined &&
      exploreState.tdd.chartType !== URLStateDefaultTDDChartType)
  ) {
    searchParams.set(
      "tdd.ct",
      ToURLParamTDDChartMap[exploreState.tdd.chartType] ?? "",
    );
  }

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
  return searchParams;
}

function toPivotUrlParams(
  exploreState: MetricsExplorerEntity,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = exploreState.pivot.rows.dimension.map(mapPivotEntry);
  if (!arrayOrderedEquals(rows, preset.pivotRows ?? [])) {
    searchParams.set("p.r", rows.join(","));
  }

  const cols = [
    ...exploreState.pivot.columns.dimension.map(mapPivotEntry),
    ...exploreState.pivot.columns.measure.map(mapPivotEntry),
  ];
  if (!arrayOrderedEquals(cols, preset.pivotCols ?? [])) {
    searchParams.set("p.c", cols.join(","));
  }

  // TODO: other fields
  return searchParams;
}
