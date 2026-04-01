import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const { organizationPermissions } = await parent();

  if (
    !organizationPermissions.manageOrg &&
    !organizationPermissions.manageProjects
  ) {
    throw error(403, "You do not have permission to access the admin console");
  }
};
