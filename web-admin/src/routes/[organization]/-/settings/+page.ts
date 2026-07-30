import { getSingleUseUrlParam } from "@rilldata/web-admin/features/navigation/getSingleUseUrlParam";
import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ url, params, parent }) => {
  const { organizationPermissions } = await parent();

  // Users with only billing access (e.g. a billing portal admin) cannot see the general settings.
  if (!organizationPermissions?.manageOrg) {
    throw redirect(307, `/${params.organization}/-/settings/billing`);
  }

  const showUpgradeDialog = !!getSingleUseUrlParam(
    url as URL,
    "upgrade",
    "rill:app:showUpgrade",
  );
  return {
    showUpgradeDialog,
  };
};
