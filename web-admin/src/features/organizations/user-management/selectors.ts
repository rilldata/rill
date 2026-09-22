import {
  createAdminServiceListOrganizationInvitesInfinite,
  createAdminServiceListOrganizationMemberUsersInfinite,
  createAdminServiceListOrganizationMemberUsergroups,
  getAdminServiceListOrganizationMemberUsergroupsQueryOptions,
  getAdminServiceListUsergroupsForOrganizationAndUserQueryOptions,
  createAdminServiceListOrganizationMemberUsergroupsInfinite,
  getAdminServiceListOrganizationMemberUsersInfiniteQueryOptions,
} from "@rilldata/web-admin/client";
import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
import { createInfiniteQuery, createQuery } from "@tanstack/svelte-query";
import { type Readable, derived, readable } from "svelte/store";

const PAGE_SIZE = 50;

export type UserGroupForUsersInOrg = {
  id: string;
  name: string;
  count: number;
};

export function getUserGroupsForUsersInOrg(
  organization: string,
  userId: string,
  enabledStore: Readable<boolean>,
): Readable<{
  isPending: boolean;
  error: unknown;
  data: UserGroupForUsersInOrg[];
}> {
  const orgUserGroupsQuery = createQuery(
    getAdminServiceListOrganizationMemberUsergroupsQueryOptions(organization, {
      pageSize: PAGE_SIZE,
      includeCounts: true,
    }),
  );
  const userGroupsForUserQuery = createQuery(
    derived(enabledStore, (enabled) =>
      getAdminServiceListUsergroupsForOrganizationAndUserQueryOptions(
        organization,
        {
          userId,
        },
        {
          query: {
            enabled: !!userId && enabled,
          },
        },
      ),
    ),
  );

  // TODO: use combine query once it supports derived options as arguments.
  return derived(
    [orgUserGroupsQuery, userGroupsForUserQuery],
    ([orgUserGroupsResp, userGroupsForUserResp]) => {
      const isPending =
        orgUserGroupsResp.isPending || userGroupsForUserResp.isPending;
      const error = orgUserGroupsResp.error ?? userGroupsForUserResp.error;

      const nonManagedGroups =
        userGroupsForUserResp.data?.usergroups?.filter((g) => !g.managed) ?? [];
      const groups = nonManagedGroups.map((g) => {
        const orgGroup = orgUserGroupsResp.data?.members?.find(
          (m) => m.groupId === g.groupId,
        );
        return {
          id: g.groupId ?? "",
          name: g.groupName ?? "",
          count: orgGroup?.usersCount ?? 0,
        };
      });

      return { isPending, error, data: groups };
    },
  );
}

const INFINITE_PAGE_SIZE = 50;

export type OrgUserMemberFilters = {
  organization: string;
  guestOnly: boolean;
  searchText?: string;
  role?: string;
  enabled?: boolean;
};

export function getOrgUserMembersQueryOptions({
  organization,
  guestOnly,
  searchText = "",
  role,
  enabled = true,
}: OrgUserMemberFilters) {
  return getAdminServiceListOrganizationMemberUsersInfiniteQueryOptions(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
      role: guestOnly ? OrgUserRoles.Guest : role,
      // Preserve literal, case-insensitive substring matching with SQL ILIKE.
      searchPattern: searchText
        ? `%${searchText.replace(/[\\%_]/g, "\\$&")}%`
        : undefined,
      includeCounts: true,
    },
    {
      query: {
        enabled: !!organization && enabled,
        initialPageParam: undefined,
        getNextPageParam: (lastPage) => lastPage.nextPageToken || undefined,
        // Keep the table mounted during search/role changes, but never show
        // another organization's users as placeholder data.
        placeholderData: (previousData, previousQuery) =>
          previousQuery?.queryKey[1] === `/v1/orgs/${organization}/members`
            ? previousData
            : undefined,
      },
    },
  );
}

export function getOrgUserMembers(
  filters: OrgUserMemberFilters | Readable<OrgUserMemberFilters>,
) {
  const filtersStore = "subscribe" in filters ? filters : readable(filters);
  return createInfiniteQuery(
    derived(filtersStore, getOrgUserMembersQueryOptions),
  );
}

export function getOrgAdminMembers(organization: string) {
  return createAdminServiceListOrganizationMemberUsersInfinite(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
      role: OrgUserRoles.Admin,
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

export function getOrgUsergroupsInfinite(organization: string) {
  return createAdminServiceListOrganizationMemberUsergroupsInfinite(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
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

function getOrgUsergroups(organization: string) {
  return createAdminServiceListOrganizationMemberUsergroups(organization, {
    pageSize: PAGE_SIZE,
    includeCounts: true,
  });
}
