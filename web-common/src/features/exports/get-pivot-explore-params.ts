import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";

export function getPivotExploreParams(
  urlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  const newUrlParams = new URLSearchParams(urlParams);
  const view = newUrlParams.get(ExploreStateURLParams.WebView);
  switch (view) {
    case "pivot":
      break;

    case "explore":
    default: {
      newUrlParams.set(ExploreStateURLParams.WebView, "pivot");

      const dim =
        newUrlParams.get(ExploreStateURLParams.ExpandedDimension) ??
        exploreSpec.dimensions?.[0] ??
        "";
      newUrlParams.set(ExploreStateURLParams.PivotRows, dim);

      const measures =
        newUrlParams.get(ExploreStateURLParams.VisibleMeasures) ??
        exploreSpec.measures?.join(",") ??
        "";
      newUrlParams.set(ExploreStateURLParams.PivotColumns, measures);

      break;
    }

    case "tdd": {
      newUrlParams.set(ExploreStateURLParams.WebView, "pivot");

      // TODO: time grain

      const dim =
        newUrlParams.get(ExploreStateURLParams.ComparisonDimension) ??
        exploreSpec.dimensions?.[0] ??
        "";
      newUrlParams.set(ExploreStateURLParams.PivotRows, dim);

      const measure =
        newUrlParams.get(ExploreStateURLParams.ExpandedMeasure) ??
        exploreSpec.measures?.[0] ??
        "";
      newUrlParams.set(ExploreStateURLParams.PivotColumns, measure);

      break;
    }
  }

  return newUrlParams.toString();
}
