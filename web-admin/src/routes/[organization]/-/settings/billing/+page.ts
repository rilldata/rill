import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

const ShowUpgradeKey = "rill:app:showUpgrade";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  if (url.searchParams.has("upgrade")) {
    try {
      localStorage.setItem(ShowUpgradeKey, "true");
    } catch {
      // no-op
    }
    throw redirect(307, `/${organization}/-/settings/billing`);
  }

  const showUpgrade = !!localStorage.getItem(ShowUpgradeKey);
  localStorage.removeItem(ShowUpgradeKey);
  return {
    organization,
    showUpgrade,
  };
};
