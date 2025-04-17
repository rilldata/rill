import { isParamCommonButDifferentMeaning } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
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
  searchParams.forEach((value, key: ExploreStateURLParams) => {
    const defaultValue = defaultUrlParams.get(key);

    if (
      // if there is no default value then skip setting if the value is empty then skip adding it.
      (defaultValue === null && value === "") ||
      // else if there is a default value,
      (defaultValue !== null &&
        // make sure it is not one of the common param name but different meaning
        !isParamCommonButDifferentMeaning(currentView, defaultView, key) &&
        // and values match then skip adding it
        value === defaultValue)
    ) {
      return;
    }

    strippedUrlParams.set(key, value);
  });
  return strippedUrlParams;
}
