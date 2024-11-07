import {
  type Bookmarks,
  fetchBookmarks,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { fetchExploreSpec } from "@rilldata/web-admin/features/dashboards/selectors";
import {
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { project, runtime } = await parent();

  const exploreName = params.dashboard;

  depends(exploreName, "explore");

  try {
    const { explore, metricsView, basePreset } = await fetchExploreSpec(
      runtime?.instanceId,
      exploreName,
    );

    // used to merge home bookmark to url state
    const bookmarks = await fetchBookmarks(
      project.id,
      exploreName,
      metricsView.metricsView?.state?.validSpec,
      explore.explore?.state?.validSpec,
    );

    return {
      explore,
      metricsView,
      basePreset,
      bookmarks,
    };
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      basePreset: <V1ExplorePreset>{},
      bookmarks: <Bookmarks>{},
    };
  }
};
