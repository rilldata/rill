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
    // Skip "cleaning" time grain param if time range is not also equivalent
    if (key === ExploreStateURLParams.TimeGrain) {
      if (
        searchParams.get(ExploreStateURLParams.TimeRange) !==
        defaultUrlParams.get(ExploreStateURLParams.TimeRange)
      ) {
        cleanedParams.set(key, value);
      }
    }

    const defaultValue = defaultUrlParams.get(key);

    const hasNoDefaultAndValueIsEmpty = defaultValue === null && value === "";
    const hasDefaultAndIsUnmodified =
      defaultValue !== null &&
      // make sure it is not one of the params that have the same name but different meaning
      !isParamCommonButDifferentMeaning(currentView, defaultView, key) &&
      paramsAreEqual(key, value, defaultValue);

    if (hasNoDefaultAndValueIsEmpty || hasDefaultAndIsUnmodified) {
      return;
    }

    cleanedParams.set(key, value);
  });
  return cleanedParams;
}

function paramsAreEqual(
  key: string,
  value: string,
  defaultValue: string,
): boolean {
  if (key === ExploreStateURLParams.Filters) {
    // Normalize array bracket syntax: IN (['x','y']) → IN ('x','y')
    // The Go backend wraps IN values in arrays while the UI uses individual strings.
    // Both are semantically identical, so normalize before comparing.
    return normalizeFilterParam(value) === normalizeFilterParam(defaultValue);
  }
  return value === defaultValue;
}

/**
 * Normalizes filter param syntax for comparison.
 * Strips array brackets from IN expressions and normalizes IN LIST → IN.
 */
function normalizeFilterParam(filter: string): string {
  return filter
    .replace(/\bNOT IN LIST\b/gi, "NIN")
    .replace(/\bIN LIST\b/gi, "IN")
    .replace(/\(\[([^\]]*)\]\)/g, "($1)");
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
