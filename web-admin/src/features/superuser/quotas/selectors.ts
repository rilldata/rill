// Quota-related queries and mutations for the superuser console
import {
  createAdminServiceGetOrganization,
  createAdminServiceSudoUpdateOrganizationQuotas,
} from "@rilldata/web-admin/client";

export function getOrgForQuotas(org: string) {
  return createAdminServiceGetOrganization(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createUpdateOrgQuotasMutation() {
  return createAdminServiceSudoUpdateOrganizationQuotas();
}
