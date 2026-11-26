import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-resolver-expression-to-v1-expression.ts";
import type { MetricsResolverQueryMapperArgs } from "@rilldata/web-common/features/explore-mappers/url-mappers/types.ts";
import { parseTimeRangeFromFilters } from "@rilldata/web-common/features/explore-mappers/parse-time-range-from-filters.ts";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { FromLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers.ts";
import {
  ExploreUrlWebView,
  ToURLParamSortTypeMap,
  ToURLParamTimeDimensionMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";
import { isAbsoluteTimeRange } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { TokenToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains.ts";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import {
  convertAbsoluteTimeRangeToRillTime,
  convertAbsoluteTimeRangeToUrlParam,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import type {
  Dimension,
  Measure,
  Schema as MetricsResolverQuery,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * Maps a metrics resolver query to explore url params.
 * The specs are not needed: the params are validated when the explore is built from the url.
 *
 * @param timeDimension The metrics view's time dimension, if already loaded by the caller.
 * Used to move a filter on the time dimension to the time range.
 * When missing, the dimension the query floors to a time grain is used.
 */
export function convertMetricsResolverQueryToUrlParams(
  args: MetricsResolverQueryMapperArgs,
  timeDimension?: string,
) {
  const searchParams = new URLSearchParams();
  const query = args.query;
  // The runtime evaluates a query without a zone in UTC.
  const timeZone = query.time_zone ?? "UTC";

  // Dimensions with a time floor are the time dimension at a grain; they are shown as time columns or charts.
  const dimensions = query.dimensions ?? [];
  const exploreDimensions = dimensions
    .filter((d) => !d.compute?.time_floor)
    .map((d) => d.name);
  const timeDimensions = dimensions.filter((d) => !!d.compute?.time_floor);
  if (exploreDimensions.length) {
    searchParams.set(
      ExploreStateURLParams.VisibleDimensions,
      exploreDimensions.join(","),
    );
  }

  const measureNames =
    query.measures?.filter((m) => !m.compute).map((m) => m.name) ?? [];
  if (measureNames.length) {
    searchParams.set(
      ExploreStateURLParams.VisibleMeasures,
      measureNames.join(","),
    );
  }

  copyParamsToTarget(
    convertResolverTimeRangesToUrlParams(query, args, timeZone),
    searchParams,
  );

  copyParamsToTarget(
    convertResolverWhereToUrlParams(
      query,
      timeDimension ?? timeDimensions[0]?.compute?.time_floor?.dimension,
      timeZone,
    ),
    searchParams,
  );

  // Absolute time ranges are written as wall-clock times in the zone, so it is pinned for them.
  const timeRangeParam = searchParams.get(ExploreStateURLParams.TimeRange);
  if (
    query.time_zone ||
    (timeRangeParam && isAbsoluteTimeRange(timeRangeParam))
  ) {
    searchParams.set(ExploreStateURLParams.TimeZone, timeZone);
  }

  copyParamsToTarget(
    convertResolverQueryToActivePageUrlParams(
      query,
      exploreDimensions,
      measureNames,
      timeDimensions,
    ),
    searchParams,
  );

  return searchParams;
}

function convertResolverTimeRangesToUrlParams(
  query: MetricsResolverQuery,
  args: MetricsResolverQueryMapperArgs,
  timeZone: string,
) {
  const searchParams = new URLSearchParams();

  if (args.resolvedTimeRange) {
    searchParams.set(
      ExploreStateURLParams.TimeRange,
      convertAbsoluteTimeRangeToRillTime(
        args.resolvedTimeRange.start!,
        args.resolvedTimeRange.end!,
        timeZone,
      ),
    );
  } else {
    searchParams.set(
      ExploreStateURLParams.TimeRange,
      convertResolverTimeRangeToUrlParam(query.time_range, timeZone),
    );
  }

  if (args.resolvedComparisonTimeRange) {
    searchParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      convertAbsoluteTimeRangeToUrlParam(
        args.resolvedComparisonTimeRange.start!,
        args.resolvedComparisonTimeRange.end!,
      ),
    );
  } else if (query.comparison_time_range) {
    // The comparison param takes an absolute range as UTC instants.
    const { start, end } = query.comparison_time_range;
    searchParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      start && end
        ? convertAbsoluteTimeRangeToUrlParam(start, end)
        : convertResolverTimeRangeToUrlParam(
            query.comparison_time_range,
            timeZone,
          ),
    );
  }

  return searchParams;
}

function convertResolverTimeRangeToUrlParam(
  timeRange: TimeRange | undefined,
  timeZone: string,
) {
  // Default to "All Time" when no time range is specified
  if (!timeRange) return TimeRangePreset.ALL_TIME;

  if (timeRange.start && timeRange.end) {
    return convertAbsoluteTimeRangeToRillTime(
      timeRange.start,
      timeRange.end,
      timeZone,
    );
  } else if (timeRange.expression) {
    return timeRange.expression;
  } else if (timeRange.iso_duration) {
    return timeRange.iso_duration;
  }

  // Fallback to all-time
  return TimeRangePreset.ALL_TIME;
}

/**
 * Sets the filter param from the query's where filter.
 * A filter on the time dimension replaces the time range, since explore filters cannot hold one.
 */
function convertResolverWhereToUrlParams(
  query: MetricsResolverQuery,
  timeDimension: string | undefined,
  timeZone: string,
) {
  const searchParams = new URLSearchParams();
  let where = mapResolverExpressionToV1Expression(query.where);
  if (!where) return searchParams;

  if (timeDimension) {
    const timeRangeParam = parseTimeRangeFromFilters(
      where,
      timeDimension,
      timeZone,
    );
    if (timeRangeParam) {
      searchParams.set(ExploreStateURLParams.TimeRange, timeRangeParam);
      // The range is written as wall-clock times in the zone, so it is pinned.
      searchParams.set(ExploreStateURLParams.TimeZone, timeZone);
      // Wrapped so a lone condition on the time dimension is removed as well;
      // only the children of a joiner are filtered.
      where = filterIdentifiers(
        createAndExpression([where]),
        (_, i) => i !== timeDimension,
      );
    }
  }

  const filterParam = convertExpressionToFilterParam(where);
  if (filterParam) {
    searchParams.set(ExploreStateURLParams.Filters, filterParam);
  }
  return searchParams;
}

/**
 * Picks the view that best shows the query:
 * time dimension details for a single measure over time, optionally split by a dimension;
 * the dimension table for a single dimension;
 * and the pivot for anything with more dimensions.
 */
function convertResolverQueryToActivePageUrlParams(
  query: MetricsResolverQuery,
  exploreDimensions: string[],
  measureNames: string[],
  timeDimensions: Dimension[],
) {
  const searchParams = new URLSearchParams();

  const showTDD =
    timeDimensions.length === 1 &&
    measureNames.length === 1 &&
    exploreDimensions.length <= 1;
  const showDimensionTable = !showTDD && exploreDimensions.length === 1;
  const showPivot = timeDimensions.length > 1 || exploreDimensions.length > 1;

  if (showTDD) {
    searchParams.set(
      ExploreStateURLParams.WebView,
      ExploreUrlWebView.TimeDimension,
    );
    searchParams.set(ExploreStateURLParams.ExpandedMeasure, measureNames[0]);
    if (exploreDimensions[0]) {
      searchParams.set(
        ExploreStateURLParams.ComparisonDimension,
        exploreDimensions[0],
      );
    }
    // Selected time grain is used in TDD's pivot table at the bottom.
    const grain = timeDimensions[0].compute?.time_floor?.grain;
    if (grain) searchParams.set(ExploreStateURLParams.TimeGrain, grain);
  } else if (showDimensionTable) {
    searchParams.set(ExploreStateURLParams.WebView, ExploreUrlWebView.Explore);
    searchParams.set(
      ExploreStateURLParams.ExpandedDimension,
      exploreDimensions[0],
    );
    copyParamsToTarget(
      convertResolverSortToLeaderboardUrlParams(
        query.measures ?? [],
        query,
        measureNames,
      ),
      searchParams,
    );
  } else if (showPivot) {
    searchParams.set(ExploreStateURLParams.WebView, ExploreUrlWebView.Pivot);
    copyParamsToTarget(
      convertResolverQueryToPivotUrlParams(
        query,
        exploreDimensions,
        measureNames,
        timeDimensions,
      ),
      searchParams,
    );
  } else {
    copyParamsToTarget(
      convertResolverSortToLeaderboardUrlParams(
        query.measures ?? [],
        query,
        measureNames,
      ),
      searchParams,
    );
  }

  return searchParams;
}

function convertResolverSortToLeaderboardUrlParams(
  measures: Measure[],
  query: MetricsResolverQuery,
  measureNames: string[],
) {
  const searchParams = new URLSearchParams();
  const sortField = query.sort?.[0];
  if (!sortField) return searchParams;

  const measure = measures.find((m) => m.name === sortField.name);
  if (!measure) return searchParams;
  const { name, type } = getMeasureNameAndType(measure);
  if (!measureNames.includes(name)) return searchParams;

  searchParams.set(ExploreStateURLParams.SortBy, name);
  searchParams.set(
    ExploreStateURLParams.SortDirection,
    sortField.desc ? "DESC" : "ASC",
  );
  const sortType = ToURLParamSortTypeMap[FromLegacySortTypeMap[type]];
  if (sortType) searchParams.set(ExploreStateURLParams.SortType, sortType);
  return searchParams;
}

function getMeasureNameAndType(measure: Measure) {
  if (measure.compute?.comparison_delta?.measure) {
    return {
      name: measure.compute.comparison_delta.measure,
      type: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
    };
  }

  if (measure.compute?.comparison_ratio?.measure) {
    return {
      name: measure.compute.comparison_ratio.measure,
      type: DashboardState_LeaderboardSortType.DELTA_PERCENT,
    };
  }

  if (measure.compute?.percent_of_total?.measure) {
    return {
      name: measure.compute.percent_of_total.measure,
      type: DashboardState_LeaderboardSortType.PERCENT,
    };
  }

  return {
    name: measure.name,
    type: DashboardState_LeaderboardSortType.VALUE,
  };
}

function convertResolverQueryToPivotUrlParams(
  query: MetricsResolverQuery,
  exploreDimensions: string[],
  measureNames: string[],
  timeDimensions: Dimension[],
) {
  const searchParams = new URLSearchParams();

  const toTimeColumn = (d: Dimension) => {
    const grain = d.compute?.time_floor?.grain;
    if (!grain) return undefined;
    return ToURLParamTimeDimensionMap[TokenToV1TimeGrain[grain]];
  };

  const columns = [
    ...timeDimensions.map(toTimeColumn),
    ...exploreDimensions,
    ...measureNames,
  ].filter(Boolean) as string[];
  searchParams.set(ExploreStateURLParams.PivotRows, "");
  searchParams.set(ExploreStateURLParams.PivotColumns, columns.join(","));
  searchParams.set(ExploreStateURLParams.PivotTableMode, "flat");

  const sort = query.sort?.find(
    (s) =>
      measureNames.includes(s.name) ||
      exploreDimensions.includes(s.name) ||
      timeDimensions.some((td) => td.name === s.name),
  );
  if (sort) {
    const timeDimension = timeDimensions.find((td) => td.name === sort.name);
    const sortBy = timeDimension ? toTimeColumn(timeDimension) : sort.name;
    if (sortBy) {
      searchParams.set(ExploreStateURLParams.SortBy, sortBy);
      searchParams.set(
        ExploreStateURLParams.SortDirection,
        sort.desc ? "DESC" : "ASC",
      );
    }
  }

  return searchParams;
}
