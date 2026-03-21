// web-admin/src/features/admin/billing/selectors.ts
import {
  adminServiceGetPaymentsPortalURL,
  createAdminServiceSudoExtendTrial,
  createAdminServiceSudoTriggerBillingRepair,
  createAdminServiceSudoDeleteOrganizationBillingIssue,
  createAdminServiceSudoUpdateOrganizationBillingCustomer,
  createAdminServiceListOrganizationBillingIssues,
} from "@rilldata/web-admin/client";

export async function getBillingSetupURL(org: string): Promise<string> {
  const resp = await adminServiceGetPaymentsPortalURL(org, {
    setup: true,
    superuserForceAccess: true,
  });
  return resp.url ?? "";
}

export function createExtendTrialMutation() {
  return createAdminServiceSudoExtendTrial();
}

export function createBillingRepairMutation() {
  return createAdminServiceSudoTriggerBillingRepair();
}

export function createDeleteBillingIssueMutation() {
  return createAdminServiceSudoDeleteOrganizationBillingIssue();
}

export function createSetBillingCustomerMutation() {
  return createAdminServiceSudoUpdateOrganizationBillingCustomer();
}

export function getBillingIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}
