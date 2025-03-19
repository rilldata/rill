import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { dashboard: exploreName }, parent }) => {
  const { runtime } = await parent();

  // Get the Explore resource
  const {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
  } = await fetchExploreSpec(runtime.instanceId, exploreName).catch((e) => {
    console.error(e);
    throw error(404, "Unable to find dashboard");
  });

  return {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
  };
};
