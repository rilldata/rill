import { isParamCommonButDifferentMeaning } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

/**
 * Removes any params that are equal to default param value.
 * If there is no default param then we remove params with empty value.
 *
 * Right now the defaults are either the rill opinionated defaults or from yaml config.
 */
export function stripDefaultOrEmptyUrlParams(
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
  searchParams.forEach((value, key: ExploreStateURLParams) => {
    const defaultValue = defaultUrlParams.get(key);

    const hasNoDefaultAndValueIsEmpty = defaultValue === null && value === "";
    const hasDefaultAndIsUnmodified =
      defaultValue !== null &&
      // make sure it is not one of the params that have the same name but different meaning
      !isParamCommonButDifferentMeaning(currentView, defaultView, key) &&
      value === defaultValue;

    if (hasNoDefaultAndValueIsEmpty || hasDefaultAndIsUnmodified) {
      return;
    }

    strippedUrlParams.set(key, value);
  });
  return strippedUrlParams;
}

export function mergeDefaultUrlParams(
  searchParams: URLSearchParams,
  defaultUrlParams: URLSearchParams,
) {
  const finalUrlParams = new URLSearchParams(searchParams);
  defaultUrlParams.forEach((value, key) => {
    if (searchParams.has(key)) return;
    finalUrlParams.set(key, value);
  });
  return finalUrlParams;
}
