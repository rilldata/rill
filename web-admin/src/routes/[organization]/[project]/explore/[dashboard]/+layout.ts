import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { project, runtime } = await parent();

  const exploreName = params.dashboard;

  depends(exploreName, "explore");

  try {
    const { explore, metricsView, defaultExplorePreset } =
      await fetchExploreSpec(runtime?.instanceId, exploreName);

    // used to merge home bookmark to url state
    const bookmarks = await fetchBookmarks(project.id, exploreName);

    return {
      explore,
      metricsView,
      defaultExplorePreset,
      homeBookmark: bookmarks.find(isHomeBookmark),
    };
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      defaultExplorePreset: <V1ExplorePreset>{},
      homeBookmark: undefined,
    };
  }
};
