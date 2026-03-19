// web-admin/src/features/admin/quotas/selectors.ts
import {
  createAdminServiceGetOrganization,
  createAdminServiceSudoUpdateOrganizationQuotas,
  createAdminServiceSudoUpdateUserQuotas,
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

export function createUpdateUserQuotasMutation() {
  return createAdminServiceSudoUpdateUserQuotas();
}
