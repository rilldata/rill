import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { project, runtime } = await parent();

  const { dashboard: exploreName } = params;

  depends(exploreName, "explore");

  let explore: V1Resource | undefined;
  let metricsView: V1Resource | undefined;
  let defaultExplorePreset: V1ExplorePreset | undefined;
  let exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity> = {};
  try {
    ({
      explore,
      metricsView,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
    } = await fetchExploreSpec(runtime?.instanceId, exploreName));
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      defaultExplorePreset: <V1ExplorePreset>{},
      exploreStateFromYAMLConfig,
    };
  }

  const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore.explore?.state?.validSpec ?? {};

  let homeBookmarkExploreState: Partial<MetricsExplorerEntity> | undefined =
    undefined;
  try {
    const bookmarks = await fetchBookmarks(project.id, exploreName);
    const homeBookmark = bookmarks.find(isHomeBookmark);

    if (homeBookmark) {
      homeBookmarkExploreState = getDashboardStateFromUrl(
        homeBookmark.data ?? "",
        metricsViewSpec,
        exploreSpec,
        {}, // TODO
      );
    }
  } catch {
    // TODO
  }

  return {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
    homeBookmarkExploreState,
  };
};
