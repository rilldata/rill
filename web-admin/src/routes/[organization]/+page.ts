import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { error } from "@sveltejs/kit";

export const load = async ({ parent }) => {
  const { organizationPermissions } = await parent();

  if (!organizationPermissions.readOrg) {
    throw error(403, m.route_error_no_org_permission());
  }
};
