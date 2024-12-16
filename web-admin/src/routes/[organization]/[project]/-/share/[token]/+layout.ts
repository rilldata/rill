import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent }) => {
  const { runtime } = await parent();

  try {
    const tokenData = await fetchMagicAuthToken(token);

    if (!tokenData.token?.resourceName) {
      throw new Error("Token does not have an associated resource name");
    }

    const { explore, metricsView, defaultExplorePreset } =
      await fetchExploreSpec(runtime, tokenData.token.resourceName);

    return {
      explore,
      metricsView,
      defaultExplorePreset,
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find token");
  }
};
