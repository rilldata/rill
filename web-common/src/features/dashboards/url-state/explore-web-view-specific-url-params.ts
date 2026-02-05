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
    ExploreStateURLParams.ComparisonDimension,
  ],
  tdd: [
    ExploreStateURLParams.WebView,
    ExploreStateURLParams.ExpandedMeasure,
    ExploreStateURLParams.ChartType,
    ExploreStateURLParams.Pin,
    ExploreStateURLParams.TimeGrain,
    ExploreStateURLParams.HighlightedTimeRange,
    ExploreStateURLParams.ComparisonDimension,
  ],
  pivot: [
    ExploreStateURLParams.WebView,
    ExploreStateURLParams.PivotRows,
    ExploreStateURLParams.PivotColumns,
    ExploreStateURLParams.PivotTableMode,
    ExploreStateURLParams.PivotRowLimit,
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
      ExploreStateURLParams.ComparisonDimension,
    ],
  },
  tdd: {
    explore: [
      ExploreStateURLParams.TimeGrain,
      ExploreStateURLParams.HighlightedTimeRange,
      ExploreStateURLParams.ComparisonDimension,
    ],
  },
};

// Sparse map of keys that are common by name between some but mean different things.
// TODO: build this automatically given ExploreWebViewSpecificURLParams and ExploreWebViewCommonURLParams
export const ExploreWebViewCommonNameDifferentMeaningURLParams: Partial<
  Record<
    ExploreUrlWebView,
    Partial<Record<ExploreUrlWebView, ExploreStateURLParams[]>>
  >
> = {
  explore: {
    pivot: [
      ExploreStateURLParams.SortBy,
      ExploreStateURLParams.SortType,
      ExploreStateURLParams.SortDirection,
    ],
  },
  pivot: {
    explore: [
      ExploreStateURLParams.SortBy,
      ExploreStateURLParams.SortType,
      ExploreStateURLParams.SortDirection,
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

export function isParamCommonButDifferentMeaning(
  view1: ExploreUrlWebView,
  view2: ExploreUrlWebView,
  param: ExploreStateURLParams,
) {
  return ExploreWebViewCommonNameDifferentMeaningURLParams[view1]?.[
    view2
  ]?.includes(param);
}
