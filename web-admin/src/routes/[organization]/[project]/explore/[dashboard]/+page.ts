import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset, bookmarks } =
    await parent();
  const { dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  // On the first dashboard load, if there are no URL params, redirect to the "Home" bookmark.
  if (
    bookmarks.home?.metricsEntity &&
    ![...url.searchParams.keys()].length &&
    !(exploreName in get(metricsExplorerStore).entities)
  ) {
    const newUrl = new URL(url);
    const searchParamsFromHomeBookmark = convertExploreStateToURLSearchParams(
      bookmarks.home.metricsEntity,
      exploreSpec,
      defaultExplorePreset,
    );
    mergeSearchParams(searchParamsFromHomeBookmark, newUrl.searchParams);
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
