import {
  createAdminServiceCreateWhitelistedDomain,
  createAdminServiceRemoveWhitelistedDomain,
  createAdminServiceListWhitelistedDomains,
} from "@rilldata/web-admin/client";

export function getWhitelistedDomains(org: string) {
  return createAdminServiceListWhitelistedDomains(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createAddWhitelistMutation() {
  return createAdminServiceCreateWhitelistedDomain();
}

export function createRemoveWhitelistMutation() {
  return createAdminServiceRemoveWhitelistedDomain();
}
