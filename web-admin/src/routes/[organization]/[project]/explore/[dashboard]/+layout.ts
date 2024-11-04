import type { QueryFunction } from "@rilldata/svelte-query";
import {
  type Bookmarks,
  fetchBookmarks,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetExploreQueryKey,
  runtimeServiceGetExplore,
  type V1ExplorePreset,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export const load = async ({ params, depends, parent }) => {
  const { project, prodDeployment } = await parent();
  const instanceId = prodDeployment?.runtimeInstanceId;

  const exploreName = params.dashboard;
  const queryParams = {
    name: exploreName,
  };

  depends(exploreName, "explore");

  const queryKey = getRuntimeServiceGetExploreQueryKey(instanceId, queryParams);

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetExplore>>
  > = ({ signal }) => runtimeServiceGetExplore(instanceId, queryParams, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const exploreResource = response.explore;
    const metricsViewResource = response.metricsView;

    const basePreset = getBasePreset(
      exploreResource.explore?.state?.validSpec ?? {},
      getLocalUserPreferencesState(exploreName),
    );

    const bookmarks = await fetchBookmarks(
      project.id,
      exploreName,
      metricsViewResource.metricsView?.state?.validSpec,
      exploreResource.explore?.state?.validSpec,
    );

    return {
      explore: exploreResource,
      metricsView: metricsViewResource,
      basePreset,
      bookmarks,
    };
  } catch {
    // error handled in +page.svelte for now
    // TODO: move it here
    return {
      explore: <V1Resource>{},
      metricsView: <V1Resource>{},
      basePreset: <V1ExplorePreset>{},
      bookmarks: <Bookmarks>{},
    };
  }
};
