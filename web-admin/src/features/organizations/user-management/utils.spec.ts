import {
  getAdminServiceListOrganizationInvitesInfiniteQueryKey,
  getAdminServiceListOrganizationInvitesQueryKey,
  getAdminServiceListOrganizationMemberUsergroupsInfiniteQueryKey,
  getAdminServiceListOrganizationMemberUsergroupsQueryKey,
  getAdminServiceListOrganizationMemberUsersInfiniteQueryKey,
  getAdminServiceListOrganizationMemberUsersQueryKey,
} from "@rilldata/web-admin/client";
import { QueryClient, type QueryKey } from "@tanstack/query-core";
import { describe, expect, it } from "vitest";
import {
  invalidateOrgInvites,
  invalidateOrgMemberUsers,
  invalidateOrgUsergroups,
} from "./utils";

const ORG = "acme";

// The params mirror what the user management pages query with.
// Infinite query keys are prefixed with "infinite",
// so a plain key does not match them and the paginated tables would go stale.
const cases: {
  name: string;
  invalidate: (
    queryClient: QueryClient,
    organization: string,
  ) => Promise<void[]>;
  plainKey: QueryKey;
  infiniteKey: QueryKey;
}[] = [
  {
    name: "org member users",
    invalidate: invalidateOrgMemberUsers,
    plainKey: getAdminServiceListOrganizationMemberUsersQueryKey(ORG, {
      pageSize: 50,
      includeCounts: true,
    }),
    infiniteKey: getAdminServiceListOrganizationMemberUsersInfiniteQueryKey(
      ORG,
      { pageSize: 50, includeCounts: true },
    ),
  },
  {
    name: "org invites",
    invalidate: invalidateOrgInvites,
    plainKey: getAdminServiceListOrganizationInvitesQueryKey(ORG, {
      pageSize: 50,
    }),
    infiniteKey: getAdminServiceListOrganizationInvitesInfiniteQueryKey(ORG, {
      pageSize: 50,
    }),
  },
  {
    name: "org usergroups",
    invalidate: invalidateOrgUsergroups,
    plainKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(ORG, {
      pageSize: 50,
      includeCounts: true,
    }),
    infiniteKey:
      getAdminServiceListOrganizationMemberUsergroupsInfiniteQueryKey(ORG, {
        pageSize: 50,
        includeCounts: true,
      }),
  },
];

describe.each(cases)(
  "$name invalidation",
  ({ invalidate, plainKey, infiniteKey }) => {
    it("invalidates both the plain and the infinite query", async () => {
      const queryClient = new QueryClient();
      queryClient.setQueryData(plainKey, {});
      queryClient.setQueryData(infiniteKey, { pages: [], pageParams: [] });

      await invalidate(queryClient, ORG);

      expect(queryClient.getQueryState(plainKey)?.isInvalidated).toBe(true);
      expect(queryClient.getQueryState(infiniteKey)?.isInvalidated).toBe(true);
    });

    it("does not invalidate queries for another organization", async () => {
      const queryClient = new QueryClient();
      const otherOrgKey = infiniteKey.map((part) =>
        typeof part === "string" ? part.replace(ORG, "other-org") : part,
      );
      queryClient.setQueryData(otherOrgKey, { pages: [], pageParams: [] });

      await invalidate(queryClient, ORG);

      expect(queryClient.getQueryState(otherOrgKey)?.isInvalidated).toBe(false);
    });
  },
);
