import {
  adminServiceListOrganizationBillingIssues,
  getAdminServiceGetBillingSubscriptionQueryKey,
  getAdminServiceGetOrganizationQueryKey,
  getAdminServiceListOrganizationBillingIssuesQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
  type V1BillingIssueType,
} from "@rilldata/web-admin/client";
import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";

export function invalidateBillingInfo(
  org: string,
  expectedIssueTypes: V1BillingIssueType[] = [],
) {
  return Promise.all([
    queryClient.refetchQueries(
      getAdminServiceGetBillingSubscriptionQueryKey(org),
    ),
    queryClient.invalidateQueries(getAdminServiceGetOrganizationQueryKey(org)),
    waitForUpdatedBillingIssues(org, expectedIssueTypes),
  ]);
}

const IssuesUpdateBaseWaitTime = 1000;
const IssuesUpdateWaitTimeMultiplier = 2;
const IssuesUpdateMaxTries = 10;

/**
 * Since all billing handling happen async we need to poll to make sure something changed.
 * This is an approximation and won't guarantee things actually changed.
 */
export async function waitForUpdatedBillingIssues(
  org: string,
  expectedIssueTypes: V1BillingIssueType[],
) {
  let tries = 0;
  const issuesResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(org),
    queryFn: () => adminServiceListOrganizationBillingIssues(org),
  });
  const currentBillingIssues = new Set(
    issuesResp.issues?.map((i) => i.type) ?? [],
  );
  if (expectedIssueTypes.every((t) => currentBillingIssues.has(t))) {
    // already has expected issues
    return;
  }

  const currentlyHasBlockerIssues = hasBlockerIssues(issuesResp.issues ?? []);

  while (tries < IssuesUpdateMaxTries) {
    await queryClient.refetchQueries(
      getAdminServiceListOrganizationBillingIssuesQueryKey(org),
    );

    const newIssuesResp = await queryClient.fetchQuery({
      queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(org),
      queryFn: () => adminServiceListOrganizationBillingIssues(org),
    });
    const issuesChangedFromPreviousFetch =
      newIssuesResp.issues &&
      // difference in sizes means there was a change
      (newIssuesResp.issues.length !== currentBillingIssues.size ||
        // some issue had a different type
        newIssuesResp.issues.some((i) => !currentBillingIssues.has(i.type)));
    // NOTE: if issues already changed from previous fetch we don't need to check against expectedIssueTypes
    if (issuesChangedFromPreviousFetch) {
      if (
        currentlyHasBlockerIssues !==
        hasBlockerIssues(newIssuesResp.issues ?? [])
      ) {
        // when blocker issues are either added or removed projects hibernation status changes.
        // so re-retch projects list to get updated hibernation status.
        // NOTE: right now projects are not automatically woken up when blocker issues are removed.
        void queryClient.refetchQueries(
          getAdminServiceListProjectsForOrganizationQueryKey(org),
        );
      }
      break;
    }

    await asyncWait(
      IssuesUpdateBaseWaitTime + tries * IssuesUpdateWaitTimeMultiplier,
    );
    tries++;
  }

  // re-fetch project list at the end
  return queryClient.refetchQueries(
    getAdminServiceListProjectsForOrganizationQueryKey(org),
  );
}
