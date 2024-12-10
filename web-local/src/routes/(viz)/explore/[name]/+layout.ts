import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params, depends }) => {
  const { instanceId } = get(runtime);

  const exploreName = params.name;

  depends(exploreName, "explore");

  try {
    const {
      explore,
      metricsView,
      defaultExplorePreset,
      initExploreState,
      initLoadedOutsideOfURL,
    } = await fetchExploreSpec(instanceId, exploreName);
    const initUrlSearch = initLoadedOutsideOfURL
      ? getUpdatedUrlForExploreState(
          explore.explore?.state?.validSpec ?? {},
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
  } catch (e) {
    console.error(e);
    throw error(404, "Explore not found");
  }
};
