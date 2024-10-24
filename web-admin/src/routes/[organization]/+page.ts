import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
import { checkUserAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { fetchOrganizationPermissions } from "@rilldata/web-admin/features/organizations/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { organization } }) => {
  let organizationPermissions: V1OrganizationPermissions = {};
  if (organization) {
    try {
      organizationPermissions =
        await fetchOrganizationPermissions(organization);
      return {
        organizationPermissions,
      };
    } catch (e) {
      if (e.response?.status !== 403 || (await checkUserAccess())) {
        throw error(e.response.status, "Error fetching organization");
      }
    }
  }
};
