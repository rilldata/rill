import { afterNavigate } from "$app/navigation";
import {
  invalidateAllMetricsViews,
  invalidateRuntimeQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  adminServiceGetProject,
  type V1GetProjectResponse,
} from "../../client";
import { getProjectRuntimeQueryKey } from "../projects/selectors";
import { viewAsUserStore } from "./viewAsUserStore";

/**
 * Remove the viewed as user (if any) when navigating away from the Dashboard page
 */
export function clearViewedAsUserAfterNavigate(queryClient: QueryClient) {
  afterNavigate((nav) => {
    // Only proceed if Viewing As a user on the Dashboard page
    if (!get(viewAsUserStore) || !nav.from?.params?.dashboard) return;

    // If staying within the project, set the admin's JWT
    if (!nav.to.params.dashboard && nav.to.params.project) {
      clearViewedAsUserWithinProject(
        queryClient,
        nav.to.params.organization,
        nav.to.params.project,
      );
    }

    // If leaving a project, clear the JWT outright
    if (!nav.to.params.dashboard && !nav.to.params.project) {
      clearViewedAsUserOutsideProject(queryClient);
    }
  });
}

export async function clearViewedAsUserWithinProject(
  queryClient: QueryClient,
  organization: string,
  project: string,
) {
  viewAsUserStore.set(null);

  // Get the admin's original JWT from the `GetProject` call
  const projResp = await queryClient.fetchQuery<V1GetProjectResponse>({
    queryKey: getProjectRuntimeQueryKey(organization, project),
    queryFn: () => adminServiceGetProject(organization, project),
  });
  const jwt = projResp.jwt;

  runtime.update((runtimeState) => {
    runtimeState.jwt = {
      token: jwt,
      receivedAt: Date.now(),
    };
    return runtimeState;
  });

  await invalidateAllMetricsViews(queryClient, get(runtime).instanceId);
}

async function clearViewedAsUserOutsideProject(queryClient: QueryClient) {
  viewAsUserStore.set(null);

  runtime.update((runtimeState) => {
    runtimeState.jwt = null;
    return runtimeState;
  });

  await invalidateRuntimeQueries(queryClient);
}
