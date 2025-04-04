import {
  ExploreWebViewSpecificURLParams,
  GlobalExploreURLParams,
} from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

export function cascadingUrlParamsMerge(urlParamsInOrder: URLSearchParams[]) {
  const curView =
    (getFirstParam(urlParamsInOrder, ExploreStateURLParams.WebView) as
      | ExploreUrlWebView
      | undefined) ?? ExploreUrlWebView.Explore;

  const newUrlParams = new URLSearchParams(urlParamsInOrder[0]);

  [
    ...GlobalExploreURLParams,
    ...ExploreWebViewSpecificURLParams[curView],
  ].forEach((k) => {
    const val = getFirstParam(urlParamsInOrder, k);
    if (!val) return;
    newUrlParams.set(k, val);
  });

  return newUrlParams;
}

function getFirstParam(
  urlParamsInOrder: URLSearchParams[],
  param: ExploreStateURLParams,
) {
  const params = urlParamsInOrder.find((p) => p.has(param));
  if (!params) return undefined;
  return params.get(param);
}
