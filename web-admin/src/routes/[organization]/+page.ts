import * as m from "@rilldata/web-common/paraglide/messages.js";
import { error } from "@sveltejs/kit";

export const load = async ({ parent }) => {
  const { organizationPermissions } = await parent();

  if (!organizationPermissions.readOrg) {
    throw error(403, m.route_error_no_org_permission());
  }
};
