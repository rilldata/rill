import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

export const ExploreWebViewSpecificURLParams: Record<
  ExploreUrlWebView,
  ExploreStateURLParams[]
> = {
  explore: [
    ExploreStateURLParams.WebView,
    ExploreStateURLParams.VisibleMeasures,
    ExploreStateURLParams.VisibleDimensions,
    ExploreStateURLParams.ExpandedDimension,
    ExploreStateURLParams.SortBy,
    ExploreStateURLParams.SortType,
    ExploreStateURLParams.SortDirection,
    ExploreStateURLParams.LeaderboardMeasures,
    ExploreStateURLParams.TimeGrain,
    ExploreStateURLParams.HighlightedTimeRange,
  ],
  tdd: [
    ExploreStateURLParams.WebView,
    ExploreStateURLParams.ExpandedMeasure,
    ExploreStateURLParams.ChartType,
    ExploreStateURLParams.Pin,
    ExploreStateURLParams.TimeGrain,
    ExploreStateURLParams.HighlightedTimeRange,
  ],
  pivot: [
    ExploreStateURLParams.WebView,
    ExploreStateURLParams.PivotRows,
    ExploreStateURLParams.PivotColumns,
    ExploreStateURLParams.PivotTableMode,
    ExploreStateURLParams.SortBy,
    ExploreStateURLParams.SortDirection,
  ],
};

// Sparse map of keys that are common between some but not all web views.
export const ExploreWebViewCommonURLParams: Partial<
  Record<
    ExploreUrlWebView,
    Partial<Record<ExploreUrlWebView, ExploreStateURLParams[]>>
  >
> = {
  explore: {
    tdd: [
      ExploreStateURLParams.TimeGrain,
      ExploreStateURLParams.HighlightedTimeRange,
    ],
  },
  tdd: {
    explore: [
      ExploreStateURLParams.TimeGrain,
      ExploreStateURLParams.HighlightedTimeRange,
    ],
  },
};

const ExploreURLParamsSpecificToSomeWebView = new Set(
  Object.values(ExploreWebViewSpecificURLParams).flat(),
);
// Url params that are editable across all web views.
export const GlobalExploreURLParams: ExploreStateURLParams[] = Object.values(
  ExploreStateURLParams,
).filter((k) => !ExploreURLParamsSpecificToSomeWebView.has(k));

export function copyUrlSearchParamsForView(
  srcView: ExploreUrlWebView,
  srcSearchParams: URLSearchParams,
  tarView: ExploreUrlWebView,
  tarSearchParams: URLSearchParams,
) {
  tarSearchParams.set(ExploreStateURLParams.WebView, tarView);

  // copy over global keys
  GlobalExploreURLParams.forEach((k) => {
    if (srcSearchParams.has(k)) {
      tarSearchParams.set(k, srcSearchParams.get(k)!);
    } else {
      tarSearchParams.delete(k);
    }
  });

  if (!ExploreWebViewCommonURLParams[srcView]?.[tarView]) return;
  // copy over any common keys only between the 2 views
  ExploreWebViewCommonURLParams[srcView][tarView].forEach(
    (k: ExploreStateURLParams) => {
      if (srcSearchParams.has(k)) {
        tarSearchParams.set(k, srcSearchParams.get(k)!);
      } else {
        tarSearchParams.delete(k);
      }
    },
  );
}
