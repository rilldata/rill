import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({
  params: { token, organization, project },
  parent,
}) => {
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
      initExploreState,
      initLoadedOutsideOfURL,
    } = await fetchExploreSpec(
      runtime?.instanceId,
      exploreName,
      `__${organization}__${project}`,
    );
    const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
    const exploreSpec = explore.explore?.state?.validSpec ?? {};

    if (tokenData.token?.state) {
      const exploreStateFromToken = getDashboardStateFromUrl(
        tokenData.token?.state,
        metricsViewSpec,
        exploreSpec,
        {}, // TODO
      );
      Object.assign(initExploreState, exploreStateFromToken);
    }
    const initUrlSearch =
      initLoadedOutsideOfURL || !!tokenData.token?.state
        ? getUpdatedUrlForExploreState(
            exploreSpec,
            defaultExplorePreset,
            initExploreState,
            new URLSearchParams(),
          )
        : "";
    console.log(tokenData.token?.state, initExploreState, initUrlSearch);

    return {
      explore,
      metricsView,
      defaultExplorePreset,
      initExploreState,
      initUrlSearch,
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find token");
  }
};
