// Billing-related queries and mutations for the superuser console
import {
  adminServiceGetPaymentsPortalURL,
  createAdminServiceSudoExtendTrial,
  createAdminServiceSudoDeleteOrganizationBillingIssue,
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

export function createDeleteBillingIssueMutation() {
  return createAdminServiceSudoDeleteOrganizationBillingIssue();
}

export function getBillingIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}
