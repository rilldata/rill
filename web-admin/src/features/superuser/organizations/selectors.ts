// All mutations in this file bake in `superuserForceAccess: true`. The
// superuser console routinely operates on orgs the caller isn't a member of,
// so every mutation needs the flag. Wrapping `mutateAsync` here means call
// sites just pass the business args and cannot forget.
import {
  createAdminServiceDeleteOrganization,
  createAdminServiceGetOrganization,
  createAdminServiceListOrganizationMemberUsers,
  createAdminServiceSearchProjectNames,
} from "@rilldata/web-admin/client";
import { derived } from "svelte/store";

export function getOrganization(org: string) {
  return createAdminServiceGetOrganization(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function getOrgMembers(org: string) {
  return createAdminServiceListOrganizationMemberUsers(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function getOrgProjects(org: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern: `${org}/%`, pageSize: 100 },
    {
      query: {
        enabled: org.length > 0,
        select: (data) =>
          data.names?.map((name) => {
            const slash = name.indexOf("/");
            return slash > 0 ? name.substring(slash + 1) : name;
          }) ?? [],
      },
    },
  );
}

export function createDeleteOrgMutation() {
  const mutation = createAdminServiceDeleteOrganization();
  return derived(mutation, ($m) => ({
    ...$m,
    mutateAsync: (vars: { org: string }) =>
      $m.mutateAsync({
        org: vars.org,
        params: { superuserForceAccess: true },
      }),
  }));
}

// Search for org names by searching project paths (org/project) and extracting unique org names.
// Caveat: orgs with zero projects won't appear in the results.
export function searchOrgNames(query: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern: `%${query}%/%`, pageSize: 100 },
    { query: { enabled: query.length >= 3 } },
  );
}

// Picks the best member to assume as when inspecting an org or project: an
// admin if one exists, otherwise the first member. Returns undefined when
// there is nobody we can assume as.
export function pickAssumableMember(
  members: Array<{ userEmail?: string; roleName?: string }> | undefined,
): { userEmail: string } | undefined {
  const admin = members?.find((m) => m.roleName === "admin");
  const picked = admin ?? members?.[0];
  return picked?.userEmail ? { userEmail: picked.userEmail } : undefined;
}
