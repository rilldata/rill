import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent, params }) => {
  const { organizationPermissions } = await parent();

  if (!organizationPermissions?.manageOrg) {
    throw redirect(307, `/${params.organization}/-/settings`);
  }
};
