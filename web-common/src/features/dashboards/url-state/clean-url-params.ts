import { isParamCommonButDifferentMeaning } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

/**
 * Removes any params that are equal to default param value.
 * If there is no default param then we remove params with empty value.
 *
 * Right now the defaults are either the rill opinionated defaults or from yaml config.
 */
export function cleanUrlParams(
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

  const cleanedParams = new URLSearchParams();
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

    cleanedParams.set(key, value);
  });
  return cleanedParams;
}

/**
 * Temporary fix to clean non-dashboard parameters from embed URLs.
 * When URL parameters exist, they are loaded exclusively from the URL, bypassing dashboard yaml defaults.
 * Since non-dashboard parameters are already removed in DashboardStateSync during URL state sync, we remove them here preemptively.
 *
 * TODO: Implement permanent solution for embed URLs, possibly by ignoring non-dashboard parameters in DashboardStateSync.
 */
export function cleanEmbedUrlParams(searchParams: URLSearchParams) {
  const cleanedParams = new URLSearchParams(searchParams);
  [
    "access_token",
    "instance_id",
    "kind",
    "resource",
    "runtime_host",
    "type",
    "navigation",
  ].forEach((p) => cleanedParams.delete(p));
  return cleanedParams;
}
