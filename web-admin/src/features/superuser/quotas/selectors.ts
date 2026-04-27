// All mutations in this file bake in `superuserForceAccess: true` where the
// RPC supports it. Wrapping `mutateAsync` here means call sites just pass the
// business args and cannot forget.
//
// `getOrganization` lives in ../organizations/selectors.ts — the Quotas page
// reuses it to avoid a second near-identical query selector.
import { createAdminServiceSudoUpdateOrganizationQuotas } from "@rilldata/web-admin/client";

export { getOrganization } from "@rilldata/web-admin/features/superuser/organizations/selectors";

// SudoUpdateOrganizationQuotas is already superuser-only on the server
// (`claims.Superuser(ctx)` check in the handler), so it doesn't need a
// `superuser_force_access` flag.
export function createUpdateOrgQuotasMutation() {
  return createAdminServiceSudoUpdateOrganizationQuotas();
}
