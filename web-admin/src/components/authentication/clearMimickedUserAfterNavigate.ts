import { afterNavigate } from "$app/navigation";
import { updateMimickedJWT } from "@rilldata/web-common/features/dashboards/granular-access-policies/updateMimickedJWT";
import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { viewAsUserStore } from "./viewAsUserStore";

/**
 * Remove the mimicked user (if any) when navigating away from the Dashboard page
 */
export function clearMimickedUserAfterNavigate(queryClient: QueryClient) {
  afterNavigate((nav) => {
    // Only applies if mimicking a user on the Dashboard page
    if (!get(viewAsUserStore) || !nav.from?.params?.dashboard) return;

    // If remaining within the project, set the admin's JWT
    if (!nav.to.params.dashboard && nav.to.params.project) {
      updateMimickedJWT(
        queryClient,
        nav.to.params.organization,
        nav.to.params.project,
        null
      );
    }

    // If leaving a project, clear the JWT
    if (!nav.to.params.dashboard && !nav.to.params.project) {
      viewAsUserStore.set(null);
      runtime.update((runtimeState) => {
        runtimeState.jwt = null;
        return runtimeState;
      });
      invalidateRuntimeQueries(queryClient);
    }
  });
}
