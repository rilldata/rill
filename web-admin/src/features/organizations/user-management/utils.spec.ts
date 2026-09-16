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
  buildInviteUsergroups,
  invalidateOrgInvites,
  invalidateOrgMemberUsers,
  invalidateOrgUsergroups,
  pendingInviteesMatching,
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

describe("buildInviteUsergroups", () => {
  it("omits the field when no groups are selected", () => {
    expect(buildInviteUsergroups([])).toBeUndefined();
    expect(buildInviteUsergroups(undefined)).toBeUndefined();
  });

  it("passes selected groups through unchanged", () => {
    expect(buildInviteUsergroups(["analysts", "eng"])).toEqual([
      "analysts",
      "eng",
    ]);
  });
});

describe("pendingInviteesMatching", () => {
  const invites = [
    { email: "Alice@example.com", roleName: "viewer" },
    { email: "bob@example.com", roleName: "viewer" },
    { email: "carol@other.org", roleName: "editor" },
  ];

  it("returns nothing without search text, matching the member search", () => {
    expect(pendingInviteesMatching(invites, "")).toEqual([]);
    expect(pendingInviteesMatching(invites, "   ")).toEqual([]);
    expect(pendingInviteesMatching(undefined, "a")).toEqual([]);
  });

  it("matches on a case-insensitive email prefix and flags the rows as pending", () => {
    expect(pendingInviteesMatching(invites, "al")).toEqual([
      { userEmail: "Alice@example.com", pendingAcceptance: true },
    ]);
    expect(pendingInviteesMatching(invites, "B")).toEqual([
      { userEmail: "bob@example.com", pendingAcceptance: true },
    ]);
    // A prefix match only: "example" appears in the domain, not at the start
    expect(pendingInviteesMatching(invites, "example")).toEqual([]);
  });
});
