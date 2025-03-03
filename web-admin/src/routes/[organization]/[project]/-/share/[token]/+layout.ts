import type { V1GetCurrentMagicAuthTokenResponse } from "@rilldata/web-admin/client/index.js";
import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  fetchExploreSpec,
  fetchMetricsViewSchema,
} from "@rilldata/web-common/features/explores/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent, url }) => {
  const { runtime } = await parent();

  let exploreName: string | undefined;
  let tokenData: V1GetCurrentMagicAuthTokenResponse | undefined;

  // Get the Explore name
  const isReport = url.searchParams.has("resource"); // Reports specify the resource in the URL
  if (isReport) {
    exploreName = url.searchParams.get("resource");
    // When we support reports for non-Explore resources, we'll also want to get the resource kind here
  } else {
    // Public URLs specify the resource in the token's metadata
    const tokenData = await fetchMagicAuthToken(token).catch((e) => {
      console.error(e);
      throw error(404, "Unable to find token");
    });

    if (!tokenData.token?.resourceName) {
      console.error("Token does not have an associated resource name");
      throw error(404, "Unable to find resource");
    }
    exploreName = tokenData.token.resourceName;
  }

  try {
    // Get the Explore resource
    const {
      explore,
      metricsView,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
    } = await fetchExploreSpec(runtime.instanceId, exploreName!);
    const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
    const exploreSpec = explore.explore?.state?.validSpec ?? {};

    // Get the initial state from the token
    let tokenExploreState: Partial<MetricsExplorerEntity> | undefined =
      undefined;
    if (tokenData?.token?.state) {
      const schema = await fetchMetricsViewSchema(
        runtime.instanceId,
        exploreSpec.metricsView ?? "",
      );
      tokenExploreState = getDashboardStateFromUrl(
        tokenData.token.state,
        metricsViewSpec,
        exploreSpec,
        schema,
      );
    }

    return {
      explore,
      metricsView,
      defaultExplorePreset,
      exploreStateFromYAMLConfig,
      tokenExploreState,
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find dashboard");
  }
};
