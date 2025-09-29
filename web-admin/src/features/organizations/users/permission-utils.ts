import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";

export function canManageOrgUser(
  organizationPermissions: V1OrganizationPermissions,
  role: string,
) {
  return (
    (role === OrgUserRoles.Admin && organizationPermissions.manageOrgAdmins) ||
    (role !== OrgUserRoles.Admin && organizationPermissions.manageOrgMembers)
  );
}
