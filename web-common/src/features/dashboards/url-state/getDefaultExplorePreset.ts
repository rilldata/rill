import {
  createAndExpression,
  flattenExpression,
  maybeConvertEqualityToInExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  dropUnknownFieldsFromFilter,
  getExploreFieldNames,
  type ExploreFieldNames,
} from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import {
  getDefaultTimeRange,
  getDefaultTimeZone,
} from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { ToURLParamTDDChartMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1ExploreComparisonMode,
  V1ExploreSortType,
  V1ExploreWebView,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { ALL_TIME_RANGE_ALIAS } from "../time-controls/new-time-controls";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";

export function getDefaultExplorePreset(
  explore: V1ExploreSpec,
  metricsViewSpec: V1MetricsViewSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const defaultMeasure =
    explore.defaultPreset?.measures?.[0] ?? explore.measures?.[0];

  const defaultExplorePreset: V1ExplorePreset = {
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
    where: createAndExpression([]),
    dimensionsWithInlistFilter: [],

    measures: explore.measures,
    dimensions: explore.dimensions,

    timeRange:
      explore.defaultPreset?.timeRange ||
      getDefaultTimeRange(metricsViewSpec.smallestTimeGrain, timeRangeSummary),
    timezone: getDefaultTimeZone(explore.timeZones),
    // No grain: the default is adaptive, so the grain follows the time range.
    timeGrain: "",
    comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
    compareTimeRange: "",
    comparisonDimension: "",

    exploreSortBy: defaultMeasure,
    exploreSortAsc: false,
    exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE,
    exploreExpandedDimension: "",
    exploreLeaderboardMeasures: defaultMeasure ? [defaultMeasure] : [],
    exploreLeaderboardShowContextForAllMeasures: false,

    timeDimensionMeasure: "",
    timeDimensionChartType: ToURLParamTDDChartMap[TDDChart.DEFAULT],
    timeDimensionPin: false,

    pivotCols: [],
    pivotRows: [],
    pivotSortBy: "",
    pivotSortAsc: false,
    pivotTableMode: "nest",
    pivotShowTotalsColumn: true,
    pivotShowTotalsRow: true,
    pivotTotalsRowPosition: "top",

    ...(explore.defaultPreset ?? {}),
  };
  // The spread above copied the default filter as the runtime parsed it.
  // The preset is turned into url params and filter chips, which need the normalized shape.
  defaultExplorePreset.where =
    getExploreDefaultFilter(metricsViewSpec, explore) ??
    createAndExpression([]);

  if (defaultExplorePreset.comparisonMode) {
    Object.assign(
      defaultExplorePreset,
      getDefaultComparisonFields(
        defaultExplorePreset,
        explore,
        timeRangeSummary,
      ),
    );
  }

  return defaultExplorePreset;
}

/**
 * The default filter of an explore, from `defaults.filters`, in the shape the filter chips write.
 *
 * The runtime parses the filter from Metrics SQL, which nests ANDs, keeps the values of an IN as
 * one list and keeps `=` as is. The chips flatten all of these, so the default is normalized the
 * same way. Otherwise the same filter would have two url params, and the dashboard could not tell
 * that it is showing its default.
 */
export function getExploreDefaultFilter(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreFields = getExploreFieldNames(metricsViewSpec, exploreSpec),
): V1Expression | undefined {
  const where = exploreSpec.defaultPreset?.where;
  if (!where) return undefined;

  // The spec is shared through the query cache and flattenExpression edits its argument.
  const flattened = flattenExpression(structuredClone(where));
  flattened.cond!.exprs = flattened.cond!.exprs?.map(
    maybeConvertEqualityToInExpressions,
  );

  // Only fields the metrics view still defines can be queried. See getExploreFieldNames.
  return dropFilterFieldsOutsideExplore(
    flattened,
    metricsViewSpec,
    exploreFields,
  );
}

/**
 * Drops the conditions of a filter on fields other than the given fields of an explore.
 * Returns undefined when no condition is left. The filter passed in is left as it is.
 */
export function dropFilterFieldsOutsideExplore(
  filter: V1Expression,
  metricsViewSpec: V1MetricsViewSpec,
  exploreFields: ExploreFieldNames,
): V1Expression | undefined {
  const measures = new Set(exploreFields.measures);
  const dimensions = new Set(exploreFields.dimensions);
  const { expr } = dropUnknownFieldsFromFilter(
    // A copy, since the subqueries of the filter are edited in place.
    // It is made through JSON, which unlike structuredClone also takes the rune state of a filter bar.
    JSON.parse(JSON.stringify(filter)) as V1Expression,
    getMapFromArray(
      metricsViewSpec.measures?.filter((m) => measures.has(m.name!)) ?? [],
      (m) => m.name!,
    ),
    getMapFromArray(
      metricsViewSpec.dimensions?.filter((d) => dimensions.has(d.name!)) ?? [],
      (d) => d.name!,
    ),
  );
  return expr.cond?.exprs?.length ? expr : undefined;
}

function getDefaultComparisonFields(
  defaultExplorePreset: V1ExplorePreset,
  explore: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): V1ExplorePreset {
  if (
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED ||
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    return {};
  }

  if (
    defaultExplorePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION
  ) {
    return {
      comparisonDimension:
        defaultExplorePreset.comparisonDimension || explore.dimensions?.[0],
    };
  }

  if (
    !defaultExplorePreset.timeRange ||
    defaultExplorePreset.timeRange === ALL_TIME_RANGE_ALIAS ||
    !timeRangeSummary?.min ||
    !timeRangeSummary?.max
  ) {
    return {};
  }

  let comparisonOption = defaultExplorePreset.compareTimeRange;

  if (!comparisonOption) {
    const preset = ISODurationToTimePreset(
      defaultExplorePreset.timeRange,
      true,
    );
    if (!preset) return {};

    const allTimeRange = {
      name: TimeRangePreset.ALL_TIME,
      start: new Date(timeRangeSummary.min),
      end: new Date(timeRangeSummary.max),
    };

    const timeRange = isoDurationToFullTimeRange(
      preset,
      allTimeRange.start,
      allTimeRange.end,
      defaultExplorePreset.timezone,
    );

    comparisonOption = getValidComparisonOption(
      explore.timeRanges,
      timeRange,
      undefined,
      allTimeRange,
      defaultExplorePreset.timezone,
    );
  }

  return {
    compareTimeRange: comparisonOption,
    exploreSortType:
      defaultExplorePreset.exploreSortType ??
      V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_PERCENT,
  };
}

export function getPinnedTimeZones(explore: V1ExploreSpec) {
  const yamlTimeZones = explore.timeZones;
  if (!yamlTimeZones || !yamlTimeZones.length) return DEFAULT_TIMEZONES;
  return yamlTimeZones;
}
