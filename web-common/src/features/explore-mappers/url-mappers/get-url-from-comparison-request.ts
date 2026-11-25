import { getSortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { FromLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers.ts";
import {
  ExploreUrlWebView,
  ToURLParamSortTypeMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { convertAPITimeRangesToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/utils.ts";
import type { URLTransformerArgs } from "@rilldata/web-common/features/explore-mappers/url-mappers/types.ts";
import {
  V1MetricsViewComparisonMeasureType,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";

/**
 * Opens the dimension table of the request's dimension.
 * Older alerts used comparison requests; they never saved a dashboard state, so the url is built from the request alone.
 */
export async function getUrlFromComparisonRequest({
  client,
  req,
  executionTime,
}: URLTransformerArgs<V1MetricsViewComparisonRequest>) {
  if (!req.metricsViewName) {
    throw new Error("No metricsView for request");
  }

  const searchParams = await convertAPITimeRangesToUrlParams(
    client,
    req.metricsViewName,
    req.timeRange,
    req.comparisonTimeRange,
    executionTime,
  );

  if (req.where) {
    searchParams.set(
      ExploreStateURLParams.Filters,
      convertExpressionToFilterParam(req.where),
    );
  }

  const measureNames =
    req.measures?.map((m) => m.name ?? "").filter(Boolean) ?? [];
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
    const sortType =
      ToURLParamSortTypeMap[
        FromLegacySortTypeMap[
          getSortType(
            sort.sortType ??
              V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED,
          )
        ]
      ];
    if (sortType) searchParams.set(ExploreStateURLParams.SortType, sortType);
  }

  searchParams.set(ExploreStateURLParams.WebView, ExploreUrlWebView.Explore);
  searchParams.set(
    ExploreStateURLParams.ExpandedDimension,
    req.dimension?.name ?? "",
  );

  return searchParams;
}
