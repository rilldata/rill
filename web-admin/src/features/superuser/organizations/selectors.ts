// Organization-related queries for the superuser console
import {
  createAdminServiceGetOrganization,
  createAdminServiceListOrganizationMemberUsers,
  createAdminServiceSearchProjectNames,
  createAdminServiceDeleteOrganization,
} from "@rilldata/web-admin/client";

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
        select: (data) => {
          // Extract project names from "org/project" paths
          const projects =
            data.names?.map((name) => {
              const slash = name.indexOf("/");
              return slash > 0 ? name.substring(slash + 1) : name;
            }) ?? [];
          return projects;
        },
      },
    },
  );
}

export function createDeleteOrgMutation() {
  return createAdminServiceDeleteOrganization();
}

// Search for org names by searching project paths (org/project) and extracting unique org names
export function searchOrgNames(query: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern: `%${query}%/%`, pageSize: 100 },
    { query: { enabled: query.length >= 3 } },
  );
}
