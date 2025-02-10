import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  fetchExploreSpec,
  fetchMetricsViewSchema,
} from "@rilldata/web-common/features/explores/selectors";
import { isHTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { token }, parent }) => {
  const { runtime } = await parent();

  const tokenData = await fetchMagicAuthToken(token).catch((e) => {
    if (!isHTTPError(e)) throw error(500, "Error fetching token");
    throw error(e.response.status, e.response.data.message);
  });

  if (!tokenData.token?.resourceName) {
    throw error(500, "Token does not have an associated resource name");
  }

  const exploreName = tokenData.token?.resourceName;

  const {
    explore,
    metricsView,
    defaultExplorePreset,
    exploreStateFromYAMLConfig,
  } = await fetchExploreSpec(runtime?.instanceId, exploreName).catch((e) => {
    if (!isHTTPError(e)) throw error(500, "Error fetching explore spec");
    throw error(e.response.status, e.response.data.message);
  });

  const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
  const exploreSpec = explore.explore?.state?.validSpec ?? {};

  let tokenExploreState: Partial<MetricsExplorerEntity> | undefined = undefined;
  if (tokenData.token?.state) {
    const schema = await fetchMetricsViewSchema(
      runtime?.instanceId,
      exploreSpec.metricsView ?? "",
    ).catch((e) => {
      if (!isHTTPError(e))
        throw error(500, "Error fetching metrics view schema");
      throw error(e.response.status, e.response.data.message);
    });

    tokenExploreState = getDashboardStateFromUrl(
      tokenData.token?.state,
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
};
