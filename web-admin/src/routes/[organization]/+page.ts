import { error } from "@sveltejs/kit";

export const load = async ({ parent }) => {
  const { organizationPermissions } = await parent();

  if (!organizationPermissions.readOrg) {
    throw error(403, "You do not have permission to access this organization");
  }
};
