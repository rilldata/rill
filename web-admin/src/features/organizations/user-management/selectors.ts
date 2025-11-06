import {
  createAdminServiceListOrganizationInvitesInfinite,
  createAdminServiceListOrganizationMemberUsersInfinite,
  createAdminServiceListOrganizationMemberUsergroups,
  getAdminServiceListOrganizationMemberUsergroupsQueryOptions,
  getAdminServiceListUsergroupsForOrganizationAndUserQueryOptions,
} from "@rilldata/web-admin/client";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import { createQueries } from "@tanstack/svelte-query";
import { type Readable, derived } from "svelte/store";

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
  return createQueries({
    queries: [
      getAdminServiceListOrganizationMemberUsergroupsQueryOptions(
        organization,
        {
          pageSize: PAGE_SIZE,
          includeCounts: true,
        },
      ),
      getAdminServiceListUsergroupsForOrganizationAndUserQueryOptions(
        organization,
        {
          userId,
        },
        {
          query: {
            enabled: !!userId,
          },
        },
      ),
    ],
    combine: ([allOrgGroupsResp, groupsForUserResp]) => {
      const isPending =
        allOrgGroupsResp.isPending || groupsForUserResp.isPending;
      const error = allOrgGroupsResp.error ?? groupsForUserResp.error;

      const nonManagedGroups =
        groupsForUserResp.data?.usergroups?.filter((g) => !g.managed) ?? [];
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
  });
}

const INFINITE_PAGE_SIZE = 50;

export function getOrgUserMembers({
  organization,
  guestOnly,
}: {
  organization: string;
  guestOnly: boolean;
}) {
  return createAdminServiceListOrganizationMemberUsersInfinite(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
      role: guestOnly ? OrgUserRoles.Guest : undefined,
      includeCounts: true,
    },
    {
      query: {
        getNextPageParam: (lastPage) => {
          if (lastPage.nextPageToken !== "") {
            return lastPage.nextPageToken;
          }
          return undefined;
        },
      },
    },
  );
}

export function getOrgUserInvites(organization: string) {
  return createAdminServiceListOrganizationInvitesInfinite(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
    },
    {
      query: {
        getNextPageParam: (lastPage) => {
          if (lastPage.nextPageToken !== "") {
            return lastPage.nextPageToken;
          }
          return undefined;
        },
      },
    },
  );
}

export function getUserCounts(organization: string) {
  return derived(
    [
      getOrgUserMembers({ organization, guestOnly: false }),
      getOrgUserMembers({ organization, guestOnly: true }),
      getOrgUsergroups(organization),
    ],
    ([allOrgUserMembersResp, guestOrgUserMembersResp, orgUsergroupsResp]) => {
      const allUsersCounts =
        allOrgUserMembersResp.data?.pages?.[0]?.totalCount ?? 0;
      const guestUsersCounts =
        guestOrgUserMembersResp.data?.pages?.[0]?.totalCount ?? 0;

      // Count only non-managed groups
      const groupsCount =
        orgUsergroupsResp.data?.members?.filter((g) => !g.groupManaged)
          .length ?? 0;

      return {
        membersCount: allUsersCounts - guestUsersCounts,
        guestsCount: guestUsersCounts,
        groupsCount,
      };
    },
  );
}

function getOrgUsergroups(organization: string) {
  return createAdminServiceListOrganizationMemberUsergroups(organization, {
    pageSize: PAGE_SIZE,
    includeCounts: true,
  });
}
