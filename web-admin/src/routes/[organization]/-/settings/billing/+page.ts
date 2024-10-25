import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

const ShowUpgradeKey = "rill:app:showUpgrade";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  // Save the state in localStorage and redirect to the url without it.
  // This prevents a refresh or saving the url from re-opening the page
  if (url.searchParams.has("upgrade")) {
    try {
      localStorage.setItem(ShowUpgradeKey, "true");
    } catch {
      // no-op
    }
    throw redirect(307, `/${organization}/-/settings/billing`);
  }

  const showUpgradeDialog = !!localStorage.getItem(ShowUpgradeKey);
  localStorage.removeItem(ShowUpgradeKey);
  return {
    organization,
    showUpgradeDialog,
  };
};
