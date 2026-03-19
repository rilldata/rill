// web-admin/src/features/admin/organizations/selectors.ts
import {
  createAdminServiceGetOrganization,
  createAdminServiceSudoUpdateOrganizationCustomDomain,
  createAdminServiceListOrganizationMemberUsers,
  createAdminServiceAddOrganizationMemberUser,
} from "@rilldata/web-admin/client";

export function getOrganization(org: string) {
  return createAdminServiceGetOrganization(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function getOrgAdmins(org: string) {
  return createAdminServiceListOrganizationMemberUsers(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createSetCustomDomainMutation() {
  return createAdminServiceSudoUpdateOrganizationCustomDomain();
}

export function createJoinOrgMutation() {
  return createAdminServiceAddOrganizationMemberUser();
}
