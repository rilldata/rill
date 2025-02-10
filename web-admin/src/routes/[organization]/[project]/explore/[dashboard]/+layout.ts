import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  fetchExploreSpec,
  fetchMetricsViewSchema,
} from "@rilldata/web-common/features/explores/selectors";
import { isHTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { error } from "@sveltejs/kit";

export const load = async ({ params, depends, parent }) => {
  const { user, project, runtime } = await parent();
  const { dashboard: exploreName } = params;
  depends(exploreName, "explore");

  const exploreSpecPromise = fetchExploreSpec(
    runtime?.instanceId,
    exploreName,
  ).catch((e) => {
    if (!isHTTPError(e)) throw error(500, "Error fetching explore spec");
    throw error(e.response.status, e.response.data.message);
  });

  const bookmarksPromise = user
    ? fetchBookmarks(project.id, exploreName).catch((e) => {
        if (!isHTTPError(e)) throw error(500, "Error fetching bookmarks");
        throw error(e.response.status, e.response.data.message);
      })
    : Promise.resolve([]);

  // Call `GetExplore` and `ListBookmarks` concurrently
  const [
    { explore, metricsView, defaultExplorePreset, exploreStateFromYAMLConfig },
    bookmarks,
  ] = await Promise.all([exploreSpecPromise, bookmarksPromise]);

  const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore.explore?.state?.validSpec ?? {};

  const schema = await fetchMetricsViewSchema(
    runtime.instanceId,
    exploreSpec.metricsView ?? "",
  ).catch((e) => {
    if (!isHTTPError(e)) throw error(500, "Error fetching metrics view schema");
    throw error(e.response.status, e.response.data.message);
  });

  let homeBookmarkExploreState: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  const homeBookmark = bookmarks.find(isHomeBookmark);

  if (homeBookmark) {
    try {
      homeBookmarkExploreState = getDashboardStateFromUrl(
        homeBookmark.data ?? "",
        metricsViewSpec,
        exploreSpec,
        schema,
      );
    } catch {
      throw error(500, "Error creating the home bookmark state");
    }
  }

  return {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
    homeBookmarkExploreState,
  };
};
