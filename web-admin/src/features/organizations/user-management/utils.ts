import {
  getAdminServiceListOrganizationInvitesQueryKey,
  getAdminServiceListOrganizationMemberUsersQueryKey,
  getAdminServiceListUsergroupMemberUsersQueryKey,
  type V1OrganizationPermissions,
} from "@rilldata/web-admin/client";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import type { QueryClient } from "@tanstack/query-core";

export function canManageOrgUser(
  organizationPermissions: V1OrganizationPermissions,
  role: string,
) {
  return (
    (role === OrgUserRoles.Admin && organizationPermissions.manageOrgAdmins) ||
    (role !== OrgUserRoles.Admin && organizationPermissions.manageOrgMembers)
  );
}

export async function invalidateAfterUserDelete(
  queryClient: QueryClient,
  organization: string,
) {
  await queryClient.invalidateQueries({
    queryKey: getAdminServiceListOrganizationMemberUsersQueryKey(organization),
  });

  await queryClient.invalidateQueries({
    queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
  });

  await queryClient.invalidateQueries({
    queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
      organization,
      "autogroup:users",
    ),
  });

  await queryClient.invalidateQueries({
    queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
      organization,
      "autogroup:members",
    ),
  });

  await queryClient.invalidateQueries({
    queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
      organization,
      "autogroup:guests",
    ),
  });
}
