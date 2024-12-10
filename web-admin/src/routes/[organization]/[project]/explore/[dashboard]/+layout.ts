import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { project, runtime } = await parent();

  const { organization, project: projectName, dashboard: exploreName } = params;

  depends(exploreName, "explore");

  let explore: V1Resource | undefined;
  let metricsView: V1Resource | undefined;
  let defaultExplorePreset: V1ExplorePreset | undefined;
  let initExploreState: Partial<MetricsExplorerEntity> = {};
  let initLoadedOutsideOfURL = false;
  try {
    ({
      explore,
      metricsView,
      defaultExplorePreset,
      initExploreState,
      initLoadedOutsideOfURL,
    } = await fetchExploreSpec(
      runtime?.instanceId,
      exploreName,
      `__${organization}__${projectName}`,
    ));
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      defaultExplorePreset: <V1ExplorePreset>{},
      initExploreState: {},
      initLoadedOutsideOfURL: false,
    };
  }

  const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore.explore?.state?.validSpec ?? {};

  try {
    const bookmarks = await fetchBookmarks(project.id, exploreName);
    const homeBookmark = bookmarks.find(isHomeBookmark);

    if (homeBookmark) {
      const exploreStateFromBookmark = getDashboardStateFromUrl(
        homeBookmark.data ?? "",
        metricsViewSpec,
        exploreSpec,
        {}, // TODO
      );
      Object.assign(initExploreState, exploreStateFromBookmark);
      initLoadedOutsideOfURL = true;
    }
  } catch {
    // TODO
  }
  const initUrlSearch = initLoadedOutsideOfURL
    ? getUpdatedUrlForExploreState(
        exploreSpec,
        defaultExplorePreset,
        initExploreState,
        new URLSearchParams(),
      )
    : "";

  return {
    explore,
    metricsView,
    defaultExplorePreset,
    initExploreState,
    initUrlSearch,
  };
};
