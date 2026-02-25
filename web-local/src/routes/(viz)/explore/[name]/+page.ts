import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  LOCAL_HOST,
  LOCAL_INSTANCE_ID,
} from "../../../../lib/local-runtime-config";
import { error } from "@sveltejs/kit";

export const load = async ({ params, depends }) => {
  const client = new RuntimeClient({
    host: LOCAL_HOST,
    instanceId: LOCAL_INSTANCE_ID,
  });

  const exploreName = params.name;

  depends(exploreName, "explore");

  try {
    const { explore, metricsView } = await fetchExploreSpec(
      client,
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
