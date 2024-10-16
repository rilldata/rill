import {
  adminServiceListOrganizationBillingIssues,
  getAdminServiceGetBillingSubscriptionQueryKey,
  getAdminServiceListOrganizationBillingIssuesQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";

export function invalidateBillingInfo(org: string) {
  return Promise.all([
    queryClient.refetchQueries(
      getAdminServiceGetBillingSubscriptionQueryKey(org),
    ),
    waitForUpdatedBillingIssues(org),
  ]);
}

const IssuesUpdateBaseWaitTime = 1000;
const IssuesUpdateWaitTimeMultiplier = 2;
const IssuesUpdateMaxTries = 10;

/**
 * Since all billing handling happen async we need to poll to make sure something changed.
 * This is an approximation and won't guarantee things actually changed.
 */
export async function waitForUpdatedBillingIssues(org: string) {
  let tries = 0;
  const issuesResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(org),
    queryFn: () => adminServiceListOrganizationBillingIssues(org),
  });
  const currentBillingIssues = new Set(issuesResp.issues.map((i) => i.type));

  while (tries < IssuesUpdateMaxTries) {
    await queryClient.refetchQueries(
      getAdminServiceListOrganizationBillingIssuesQueryKey(org),
    );

    const newIssuesResp = await queryClient.fetchQuery({
      queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(org),
      queryFn: () => adminServiceListOrganizationBillingIssues(org),
    });
    if (
      // difference in sizes means there was a change
      newIssuesResp.issues.length !== currentBillingIssues.size ||
      // some issue had a different type
      newIssuesResp.issues.some((i) => !currentBillingIssues.has(i.type))
    ) {
      break;
    }

    await asyncWait(
      IssuesUpdateBaseWaitTime + tries * IssuesUpdateWaitTimeMultiplier,
    );
    tries++;
  }
}
