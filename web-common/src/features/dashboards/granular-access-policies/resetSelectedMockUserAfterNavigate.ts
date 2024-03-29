import { beforeNavigate } from "$app/navigation";
import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import { updateDevJWT } from "@rilldata/web-common/features/dashboards/granular-access-policies/updateDevJWT";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

/**
 * Remove the selected mock user (if any) when navigating to a dashboard
 * (This doesn't apply when navigating from a dashboard's edit page to its view page)
 *
 * Note: It'd be better if we didn't do this. It's a hack to avoid the following bug: Navigating to
 * a dashboard where the selected mock user does not have access shows a blank page – because
 * under this scenario, the catalog entry returns a 404, and it's required to enter the top-level
 * `Dashboard.svelte` component.
 */
export function resetSelectedMockUserAfterNavigate(queryClient: QueryClient) {
  beforeNavigate(({ to, from }) => {
    if (!to?.params || !from?.params) return;

    if (
      from.params.name !== to.params.name &&
      get(selectedMockUserStore) !== null
    ) {
      updateDevJWT(queryClient, null).catch(console.error);
    }
  });
}
