import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { LOCAL_INSTANCE_ID } from "../../../../lib/local-runtime-config";
import { error } from "@sveltejs/kit";

export const load = async ({ params, depends }) => {
  const instanceId = LOCAL_INSTANCE_ID;

  const exploreName = params.name;

  depends(exploreName, "explore");

  try {
    const { explore, metricsView } = await fetchExploreSpec(
      instanceId,
      exploreName,
    );

    return {
      explore,
      metricsView,
      exploreName,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Explore not found");
  }
};
