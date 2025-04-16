import { paramValidInBothViews } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

export function stripDefaultUrlParams(
  searchParams: URLSearchParams,
  defaultUrlParams: URLSearchParams,
) {
  const currentView =
    (searchParams.get(
      ExploreStateURLParams.WebView,
    ) as ExploreUrlWebView | null) ?? ExploreUrlWebView.Explore;
  const defaultView =
    (defaultUrlParams.get(
      ExploreStateURLParams.WebView,
    ) as ExploreUrlWebView | null) ?? ExploreUrlWebView.Explore;

  const strippedUrlParams = new URLSearchParams();
  searchParams.forEach((value, key) => {
    if (
      !paramValidInBothViews(
        currentView,
        defaultView,
        key as ExploreStateURLParams,
      )
    ) {
      // If the param is not valid in both views then checking equality doesn't make sense.
      // So set it and return early.
      strippedUrlParams.set(key, value);
      return;
    }

    const defaultValue = defaultUrlParams.get(key);
    if (defaultValue !== null && value === defaultValue) {
      return;
    }
    strippedUrlParams.set(key, value);
  });
  return strippedUrlParams;
}
