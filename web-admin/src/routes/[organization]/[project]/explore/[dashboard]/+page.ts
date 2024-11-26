import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset, bookmarks } =
    await parent();
  const { dashboard: exploreName } = params;
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
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    errors.push(...errorsFromConvert);
  }

  if (
    !(exploreName in get(metricsExplorerStore).entities) &&
    bookmarks.home?.metricsEntity &&
    ![...url.searchParams.keys()].length
  ) {
    // Initial load of the dashboard.
    // Merge home bookmark to url if present and there are no params in the url
    const newUrl = new URL(url);
    const searchParamsFromHomeBookmark = convertMetricsEntityToURLSearchParams(
      bookmarks.home.metricsEntity,
      exploreSpec,
      defaultExplorePreset,
    );
    mergeSearchParams(searchParamsFromHomeBookmark, newUrl.searchParams);
    throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
  }

  return {
    partialExploreState,
    errors,
  };
};
