import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { FromURLParamViewMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Sometimes data is loaded from sources other than the url.
 * In that case update the URL to make sure the state matches the current url.
 */
export function getUpdatedUrlForExploreState(
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
  partialExploreState: Partial<MetricsExplorerEntity>,
  curSearchParams: URLSearchParams,
) {
  const newUrlSearchParams = new URLSearchParams(
    convertExploreStateToURLSearchParams(
      partialExploreState as MetricsExplorerEntity,
      exploreSpec,
      defaultExplorePreset,
    ),
  );
  curSearchParams.forEach((value, key) => {
    if (
      key === ExploreStateURLParams.WebView &&
      FromURLParamViewMap[value] === defaultExplorePreset.view
    ) {
      // ignore default view.
      // since we do not add params equal to default this will append to the end of the URL breaking the param order.
      return;
    }
    newUrlSearchParams.set(key, value);
  });
  return newUrlSearchParams.toString();
}
