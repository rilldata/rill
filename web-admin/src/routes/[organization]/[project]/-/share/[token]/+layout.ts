import { fetchExploreSpec } from "@rilldata/web-admin/features/dashboards/selectors";
import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent }) => {
  const { runtime } = await parent();

  try {
    const tokenData = await fetchMagicAuthToken(token);

    const { explore, metricsView, defaultExplorePreset } =
      await fetchExploreSpec(
        runtime?.instanceId,
        tokenData.token?.resourceName,
      );

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
