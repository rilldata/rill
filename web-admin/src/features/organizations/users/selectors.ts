import {
  createAdminServiceListOrganizationMemberUsergroups,
  createAdminServiceListUsergroupsForOrganizationAndUser,
} from "@rilldata/web-admin/client";
import { derived, type Readable } from "svelte/store";

const PAGE_SIZE = 50;

export type UserGroupForUsersInOrg = {
  id: string;
  name: string;
  count: number;
};

export function getUserGroupsForUsersInOrg(
  organization: string,
  userId: string,
): Readable<{
  isPending: boolean;
  error: unknown;
  data: UserGroupForUsersInOrg[];
}> {
  return derived(
    [
      createAdminServiceListOrganizationMemberUsergroups(organization, {
        pageSize: PAGE_SIZE,
        includeCounts: true,
      }),
      createAdminServiceListUsergroupsForOrganizationAndUser(organization, {
        userId,
      }),
    ],
    ([allOrgGroupsResp, groupsForUserResp]) => {
      const isPending =
        allOrgGroupsResp.isPending || groupsForUserResp.isPending;
      const error = allOrgGroupsResp.error ?? groupsForUserResp.error;

      const nonManagedGroups =
        groupsForUserResp.data?.usergroups.filter((g) => !g.managed) ?? [];
      const groups = nonManagedGroups.map((g) => {
        const orgGroup = allOrgGroupsResp.data?.members?.find(
          (m) => m.groupId === g.groupId,
        );
        return {
          id: g.groupId ?? "",
          name: g.groupName ?? "",
          count: orgGroup?.usersCount ?? 0,
        };
      });

      return {
        isPending,
        error,
        data: groups,
      };
    },
  );
}
