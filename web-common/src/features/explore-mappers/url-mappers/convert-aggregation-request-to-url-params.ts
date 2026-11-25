import { splitDimensionsAndMeasuresAsRowsAndColumns } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import type { URLTransformerArgs } from "@rilldata/web-common/features/explore-mappers/url-mappers/types.ts";
import {
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  V1ExploreSortType,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { convertStateParamToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/augment-url-params.ts";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";
import { convertAPITimeRangesToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/utils.ts";
import { convertQueryFilterToToplistQuery } from "@rilldata/web-common/features/explore-mappers/utils.ts";
import {
  createAndExpression,
  createSubQueryExpression,
  forEachIdentifier,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  measureHasSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  COMPARISON_VALUE,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { ephemeralDefsFromRequestMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping.ts";
import { toEphemeralMeasuresParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param.ts";
import {
  ExploreUrlWebView,
  ToURLParamSortTypeMap,
  ToURLParamTimeDimensionMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";

export async function convertAggregationRequestToUrlParams({
  client,
  req,
  executionTime,
  urlState,
  ignoreFilters,
  forceOpenPivot,
}: URLTransformerArgs<V1MetricsViewAggregationRequest>) {
  if (!req.metricsView) {
    throw new Error("No metricsView for request");
  }

  const searchParams = convertStateParamToUrlParams("explore", urlState);

  copyParamsToTarget(
    await convertAPITimeRangesToUrlParams(
      client,
      req.metricsView,
      req.timeRange,
      req.comparisonTimeRange,
      executionTime,
    ),
    searchParams,
  );

  copyParamsToTarget(
    await convertFiltersToUrlParams(client, req, ignoreFilters),
    searchParams,
  );

  // Reports and alerts embed ephemeral measure definitions in the request itself,
  // so recover them here: the saved dashboard state may be missing (older reports)
  // or absent entirely, and without the definitions the names cannot be resolved.
  const requestEphemeralMeasures = ephemeralDefsFromRequestMeasures(
    req.measures,
  );
  if (requestEphemeralMeasures?.length) {
    searchParams.set(
      ExploreStateURLParams.EphemeralMeasures,
      toEphemeralMeasuresParam(requestEphemeralMeasures),
    );
  }

  // everything after this can be loaded from the dashboard state if present
  if (urlState) return searchParams;

  if (forceOpenPivot) {
    searchParams.set(ExploreStateURLParams.WebView, ExploreUrlWebView.Pivot);
    copyParamsToTarget(
      convertAggregationRequestToPivotUrlParams(req),
      searchParams,
    );
    return searchParams;
  }

  copyParamsToTarget(
    convertAggregationRequestToExploreUrlParams(req),
    searchParams,
  );

  return searchParams;
}

async function convertFiltersToUrlParams(
  client: RuntimeClient,
  req: V1MetricsViewAggregationRequest,
  ignoreFilters: boolean | undefined,
) {
  const searchParams = new URLSearchParams();

  let where: V1Expression | undefined = undefined;

  const shouldParseWhereFilter = Boolean(!ignoreFilters && req.where);
  if (shouldParseWhereFilter) {
    // Explore state keeps dimension and measure filters collapsed into the where filter.
    where = req.where!;
  }

  const shouldParseHavingFilter = Boolean(
    !ignoreFilters &&
      req.having?.cond?.exprs?.length &&
      req.dimensions?.[0]?.name,
  );
  if (shouldParseHavingFilter) {
    const dimension = req.dimensions![0].name!;

    let havingFilter: V1Expression;
    if (exprHasComparison(req.having!)) {
      // We do not support comparison based measure filters in dashboards right now.
      // So convert it to a toplist and add `in` filter.
      havingFilter = await convertQueryFilterToToplistQuery(
        client,
        req.metricsView!,
        req,
        dimension,
      );
    } else {
      // Measure filters are stored as a subquery on the dimension within the where filter.
      havingFilter = createSubQueryExpression(
        dimension,
        getAllIdentifiers(req.having),
        req.having,
      );
    }

    where =
      mergeFilters(
        where ?? createAndExpression([]),
        createAndExpression([havingFilter]),
      ) ?? createAndExpression([]);
  }

  if (where) {
    searchParams.set(
      ExploreStateURLParams.Filters,
      convertExpressionToFilterParam(where, []),
    );
  }

  return searchParams;
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

/**
 * Opens the dimension table of the request's first dimension.
 * A request without dimensions opens the time dimension details of its first measure.
 * Explores without a time dimension fall back to the explore view during validation.
 */
function convertAggregationRequestToExploreUrlParams(
  req: V1MetricsViewAggregationRequest,
) {
  const searchParams = new URLSearchParams();

  const measureNames =
    req.measures
      ?.map((m) => m.name ?? "")
      .filter((m) => m && !measureHasSuffix(m)) ?? [];
  if (measureNames.length) {
    searchParams.set(
      ExploreStateURLParams.VisibleMeasures,
      measureNames.join(","),
    );
  }

  const sort = req.sort?.[0];
  if (sort?.name && measureNames.includes(sort.name)) {
    searchParams.set(ExploreStateURLParams.SortBy, sort.name);
    searchParams.set(
      ExploreStateURLParams.SortDirection,
      sort.desc ? "DESC" : "ASC",
    );
    searchParams.set(
      ExploreStateURLParams.SortType,
      ToURLParamSortTypeMap[V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE] ?? "",
    );
  }

  const dimensionName = req.dimensions?.[0]?.name;
  if (dimensionName) {
    searchParams.set(ExploreStateURLParams.WebView, ExploreUrlWebView.Explore);
    searchParams.set(ExploreStateURLParams.ExpandedDimension, dimensionName);
  } else {
    searchParams.set(
      ExploreStateURLParams.WebView,
      ExploreUrlWebView.TimeDimension,
    );
    searchParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      req.measures?.[0]?.name ?? "",
    );
  }

  return searchParams;
}

function convertAggregationRequestToPivotUrlParams(
  req: V1MetricsViewAggregationRequest,
) {
  const searchParams = new URLSearchParams();
  const {
    rows: reqRows,
    dimensionColumns: reqDimCols,
    measureColumns: reqMeasureCols,
  } = splitDimensionsAndMeasuresAsRowsAndColumns(req);

  const mapReqCol = (
    dim: V1MetricsViewAggregationDimension | V1MetricsViewAggregationMeasure,
  ): string | undefined => {
    if (
      "timeGrain" in dim &&
      dim.timeGrain &&
      dim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED
    ) {
      return ToURLParamTimeDimensionMap[dim.timeGrain];
    }
    return dim.name;
  };

  const rows = reqRows.map(mapReqCol).filter(Boolean) as string[];
  searchParams.set(ExploreStateURLParams.PivotRows, rows.join(","));

  const cols = reqDimCols
    .map(mapReqCol)
    .concat(reqMeasureCols.map(mapReqCol))
    .filter(Boolean) as string[];
  searchParams.set(ExploreStateURLParams.PivotColumns, cols.join(","));

  searchParams.set(
    ExploreStateURLParams.PivotTableMode,
    req.pivotOn?.length ? "nest" : "flat",
  );

  // Only the first sort is kept; it is dropped during validation if the pivot has no column for it,
  // e.g. a sort by a measure's previous value.
  const sort = req.sort?.[0];
  if (sort?.name) {
    // A request sorts a time dimension by its name or alias, while the pivot keys its column by grain.
    const dimension = req.dimensions?.find(
      (d) => d.name === sort.name || (!!d.alias && d.alias === sort.name),
    );
    const sortBy = dimension
      ? mapReqCol(dimension)
      : convertComparisonMeasureToPivotMeasures(sort.name);
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

// Map measure filter suffixes and API suffixes to pivot suffixes
// to handle the measure filter format (_delta, _delta_perc) for backwards compatibility
export function convertComparisonMeasureToPivotMeasures(measure: string) {
  if (
    measure.endsWith(COMPARISON_DELTA) ||
    measure.endsWith(COMPARISON_PERCENT) ||
    measure.endsWith(COMPARISON_VALUE)
  ) {
    return measure;
  }

  // Handle measure filter format for backwards compatibility
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
