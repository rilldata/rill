import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent }) => {
  const { explore, metricsView, basePreset, token } = await parent();
  const exploreName = token.resourceName;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];
  if (metricsViewSpec && exploreSpec) {
    const {
      partialExploreState: partialExploreStateFromUrl,
      errors: errorsFromConvert,
    } = convertURLToMetricsExplore(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      basePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
  }

  if (
    !(exploreName in get(metricsExplorerStore).entities) &&
    token.state &&
    ![...url.searchParams.keys()].length
  ) {
    // Initial load of the dashboard.
    // Merge home token state to url if present and there are no params in the url
    // convert legacy state to new readable url format
    const metricsEntity = getDashboardStateFromUrl(
      token.state,
      metricsViewSpec,
      exploreSpec,
      {}, // TODO
    );
    const newUrl = new URL(url);
    const searchParamsFromTokenState = convertMetricsEntityToURLSearchParams(
      metricsEntity,
      exploreSpec,
      basePreset,
    );
    mergeSearchParams(searchParamsFromTokenState, newUrl.searchParams);
    throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
  }

  return {
    partialExploreState,
    errors,
  };
};
