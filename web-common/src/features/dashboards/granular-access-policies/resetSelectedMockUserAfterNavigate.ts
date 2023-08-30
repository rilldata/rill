import { afterNavigate } from "$app/navigation";
import { selectedMockUserStore } from "./stores";

/**
 * Remove the selected mock user (if any) when navigating to a dashboard
 * (This doesn't apply when navigating from a dashboard's edit page to its view page)
 *
 * Note: It'd be better if we didn't do this. It's a hack to avoid the following bug: Navigating to
 * a dashboard where the selected mock user does not have access shows a blank page – because
 * under this scenario, the catalog entry returns a 401, and it's required to enter the top-level
 * `Dashboard.svelte` component.
 */
export function resetSelectedMockUserAfterNavigate() {
  afterNavigate((nav) => {
    if (nav.from.params.name !== nav.to.params.name) {
      selectedMockUserStore.set(null);
    }
  });
}
