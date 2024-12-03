import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, defaultExplorePreset, token } = await parent();
  const exploreName = token.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  // On the first dashboard load, if there are no URL params, append the token's state (in human-readable format) to the URL.
  if (
    token.state &&
    ![...url.searchParams.keys()].length &&
    !(exploreName in get(metricsExplorerStore).entities)
  ) {
    const exploreState = getDashboardStateFromUrl(
      token.state,
      metricsViewSpec,
      exploreSpec,
      {}, // TODO
    );
    const newUrl = new URL(url);
    const searchParamsFromTokenState = convertExploreStateToURLSearchParams(
      exploreState,
      exploreSpec,
      defaultExplorePreset,
    );
    mergeSearchParams(searchParamsFromTokenState, newUrl.searchParams);
    throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
  }

  // Get Explore state from URL params
  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec) {
    const {
      partialExploreState: partialExploreStateFromUrl,
      errors: errorsFromConvert,
    } = convertURLToExploreState(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
  }

  return {
    partialExploreState,
    errors,
  };
};
