import {
  getAdminServiceListOrganizationInvitesInfiniteQueryKey,
  getAdminServiceListOrganizationInvitesQueryKey,
  getAdminServiceListOrganizationMemberUsergroupsInfiniteQueryKey,
  getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  getAdminServiceListOrganizationMemberUsersInfiniteQueryKey,
  getAdminServiceListOrganizationMemberUsersQueryKey,
  getAdminServiceListUsergroupMemberUsersQueryKey,
  type V1OrganizationPermissions,
} from "@rilldata/web-admin/client";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import type { QueryClient } from "@tanstack/query-core";

// A single row of the custom attributes editor, as bound to KeyValueInput.
export type AttributeRow = { key: string; value: string };

// Builds the attributes object for an invite from key/value form rows, skipping rows with an empty key.
// Returns undefined when no attributes are set, so the request field is omitted entirely.
export function buildInviteAttributes(
  rows: AttributeRow[] | undefined,
): Record<string, string> | undefined {
  const attrs: Record<string, string> = {};
  for (const { key, value } of rows ?? []) {
    const trimmedKey = key.trim();
    if (trimmedKey) attrs[trimmedKey] = value;
  }
  return Object.keys(attrs).length > 0 ? attrs : undefined;
}

export function canManageOrgUser(
  organizationPermissions: V1OrganizationPermissions,
  role: string,
) {
  return (
    (role === OrgUserRoles.Admin && organizationPermissions.manageOrgAdmins) ||
    (role !== OrgUserRoles.Admin && organizationPermissions.manageOrgMembers)
  );
}

// The org user, invite and group lists are read through both plain and infinite queries.
// Orval caches the infinite variants under an "infinite"-prefixed query key,
// so invalidating the plain key alone leaves the paginated tables showing stale data.
// Always invalidate through these helpers rather than a single key.

export function invalidateOrgMemberUsers(
  queryClient: QueryClient,
  organization: string,
) {
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
    }),
    queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationMemberUsersInfiniteQueryKey(
          organization,
        ),
    }),
  ]);
}

export function invalidateOrgInvites(
  queryClient: QueryClient,
  organization: string,
) {
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
    }),
    queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationInvitesInfiniteQueryKey(organization),
    }),
  ]);
}

export function invalidateOrgUsergroups(
  queryClient: QueryClient,
  organization: string,
) {
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
    }),
    queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationMemberUsergroupsInfiniteQueryKey(
          organization,
        ),
    }),
  ]);
}

export async function invalidateAfterUserDelete(
  queryClient: QueryClient,
  organization: string,
) {
  await invalidateOrgMemberUsers(queryClient, organization);

  await invalidateOrgInvites(queryClient, organization);

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
