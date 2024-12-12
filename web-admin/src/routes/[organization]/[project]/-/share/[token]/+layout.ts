import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent }) => {
  const { runtime } = await parent();

  try {
    const tokenData = await fetchMagicAuthToken(token);

    if (!tokenData.token?.resourceName) {
      throw new Error("Token does not have an associated resource name");
    }

    const exploreName = tokenData.token?.resourceName;

    const {
      explore,
      metricsView,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
    } = await fetchExploreSpec(runtime?.instanceId, exploreName);
    const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
    const exploreSpec = explore.explore?.state?.validSpec ?? {};

    let initExploreState: Partial<MetricsExplorerEntity> | undefined =
      undefined;
    if (tokenData.token?.state) {
      initExploreState = getDashboardStateFromUrl(
        tokenData.token?.state,
        metricsViewSpec,
        exploreSpec,
        {}, // TODO
      );
    }

    return {
      explore,
      metricsView,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
      initExploreState,
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find token");
  }
};
