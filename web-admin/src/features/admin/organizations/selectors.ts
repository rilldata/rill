// web-admin/src/features/admin/organizations/selectors.ts
import {
  createAdminServiceGetOrganization,
  createAdminServiceListOrganizationMemberUsers,
  createAdminServiceListProjectsForOrganization,
  createAdminServiceSearchProjectNames,
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
  return createAdminServiceListProjectsForOrganization(
    org,
    {},
    { query: { enabled: org.length > 0 } },
  );
}

// Search for org names by searching project paths (org/project) and extracting unique org names
export function searchOrgNames(query: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern: `%${query}%/%`, pageSize: 100 },
    { query: { enabled: query.length >= 3 } },
  );
}
